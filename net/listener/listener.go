package listener

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

//Proto gets a connection based on content
type Proto func(io.Reader) bool

// MatchAny matches any connection.
func MatchAny() Proto {
	return func(r io.Reader) bool { return true }
}

// MatchWS only matches the HTTP GET request.
func MatchWS(strs ...string) Proto {
	pt := newPatriciaTreeString(strs...)
	return pt.matchPrefix
}

type ErrorHandler func(error) bool

var _ net.Error = ErrProtoNotMatched{}

type ErrProtoNotMatched struct {
	c net.Conn
}

func (e ErrProtoNotMatched) Error() string {
	return fmt.Sprintf("mux: proto %v is not registered with the mux",
		e.c.RemoteAddr())
}

// Temporary implements the net.Error interface.
func (e ErrProtoNotMatched) Temporary() bool { return true }

// Timeout implements the net.Error interface.
func (e ErrProtoNotMatched) Timeout() bool { return false }

type errListenerClosed string

func (e errListenerClosed) Error() string   { return string(e) }
func (e errListenerClosed) Temporary() bool { return false }
func (e errListenerClosed) Timeout() bool   { return false }

// ErrListenerClosed is returned from muxListener.Accept when the underlying
// listener is closed.
var ErrListenerClosed = errListenerClosed("listener: listener closed")

// for readability of readTimeout
var zeroTime time.Duration

type Listener struct {
	root        net.Listener
	buffSize    int
	errHandler  ErrorHandler
	closing     chan struct{}
	protos      []mux
	readTimeout time.Duration
}

func New(address string) (*Listener, error) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	return &Listener{
		root:        l,
		buffSize:    1024,
		errHandler:  func(_ error) bool { return true },
		closing:     make(chan struct{}),
		readTimeout: zeroTime,
	}, nil
}

type mux struct {
	protos []Proto
	listen muxListener
}

func (m *Listener) SetReadTimeout(t time.Duration) {
	m.readTimeout = t
}

func (m *Listener) Addr() net.Addr {
	return m.root.Addr()
}

func (m *Listener) Accept() (net.Conn, error) {
	return m.root.Accept()
}

func (m *Listener) ServeCallback(proto Proto, serve func(l net.Listener) error) {
	p := m.Proto(proto)
	go serve(p)
}

func (m *Listener) Proto(proto ...Proto) net.Listener {
	ml := muxListener{
		Listener: m.root,
		conn:     make(chan net.Conn, m.buffSize),
	}
	m.protos = append(m.protos, mux{protos: proto, listen: ml})
	return ml
}

// Serve start multiplexing the listener.
func (m *Listener) Serve() error {
	var wg sync.WaitGroup

	defer func() {
		close(m.closing)
		wg.Wait()

		for _, p := range m.protos {
			close(p.listen.conn)

			for c := range p.listen.conn {
				_ = c.Close()
			}
		}
	}()

	for {
		c, err := m.root.Accept()
		if err != nil {
			if !m.handleErr(err) {
				return err
			}
			continue
		}

		wg.Add(1)
		go m.serve(c, m.closing, &wg)
	}
}

func (m *Listener) serve(c net.Conn, donec <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	muxc := newConn(c)
	if m.readTimeout > zeroTime {
		_ = c.SetReadDeadline(time.Now().Add(m.readTimeout))
	}

	for _, p := range m.protos {
		for _, proto := range p.protos {
			matched := proto(muxc.startSniffing())
			if matched {
				muxc.doneSniffing()
				if m.readTimeout > zeroTime {
					_ = c.SetReadDeadline(time.Time{})
				}
				select {
				case p.listen.conn <- muxc:
				case <-donec:
					_ = c.Close()
				}
				return
			}
		}
	}
	_ = c.Close()
	err := ErrProtoNotMatched{c: c}
	if !m.handleErr(err) {
		_ = m.root.Close()
	}
}

func (m *Listener) HandleError(h ErrorHandler) {
	m.errHandler = h
}

func (m *Listener) handleErr(err error) bool {
	if !m.errHandler(err) {
		return false
	}

	if ne, ok := err.(net.Error); ok {
		return ne.Temporary()
	}

	return false
}

//Close closes the listener
func (m *Listener) Close() error {
	return m.root.Close()
}

type muxListener struct {
	net.Listener
	conn chan net.Conn
}

func (l muxListener) Accept() (net.Conn, error) {
	c, ok := <-l.conn
	if !ok {
		return nil, ErrListenerClosed
	}
	return c, nil
}

// Conn wraps a net.Conn and provides transparent sniffing of connection data.
type Conn struct {
	net.Conn
	buf stream
}

func newConn(c net.Conn) *Conn {
	return &Conn{
		Conn: c,
		buf:  stream{source: c},
	}
}

func (m *Conn) Read(p []byte) (int, error) {
	return m.buf.Read(p)
}

func (m *Conn) startSniffing() io.Reader {
	m.buf.reset(true)
	return &m.buf
}

func (m *Conn) doneSniffing() {
	m.buf.reset(false)
}

// Stream represents a io.Reader which can peek incoming bytes and reset back to normal.
type stream struct {
	source     io.Reader
	buffer     bytes.Buffer
	bufferRead int
	bufferSize int
	sniffing   bool
	lastErr    error
}

// Read reads data from the buffer.
func (s *stream) Read(p []byte) (int, error) {
	if s.bufferSize > s.bufferRead {
		bn := copy(p, s.buffer.Bytes()[s.bufferRead:s.bufferSize])
		s.bufferRead += bn
		return bn, s.lastErr
	} else if !s.sniffing && s.buffer.Cap() != 0 {
		s.buffer = bytes.Buffer{}
	}

	sn, sErr := s.source.Read(p)
	if sn > 0 && s.sniffing {
		s.lastErr = sErr
		if wn, wErr := s.buffer.Write(p[:sn]); wErr != nil {
			return wn, wErr
		}
	}
	return sn, sErr
}

// Reset resets the buffer.
func (s *stream) reset(snif bool) {
	s.sniffing = snif
	s.bufferRead = 0
	s.bufferSize = s.buffer.Len()
}
