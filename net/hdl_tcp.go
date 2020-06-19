package net

import (
	"errors"
	"net"
	"time"
)

//onAccept is a callback which get called when a connection is accepted
type TcpHandler func(c net.Conn, proto Proto)

// ErrServerClosed occurs when a tcp server is closed.
var ErrServerClosed = errors.New("tcp: Server closed")

func (s *Server) Serve(l net.Listener) error {
	defer l.Close()

	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-s.closing():
				return ErrServerClosed
			default:
			}

			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}

				time.Sleep(tempDelay)
				continue
			}
			return err
		}

		tempDelay = 0
		go s.TcpHandler(conn, RPC)
	}
}

// Closing gets the closing channel in a thread-safe manner.
func (s *Server) closing() <-chan bool {
	s.Lock()
	defer s.Unlock()
	return s.getClosing()
}

// Closing gets the closing channel in a non thread-safe manner.
func (s *Server) getClosing() chan bool {
	if s.Closing == nil {
		s.Closing = make(chan bool)
	}
	return s.Closing
}
