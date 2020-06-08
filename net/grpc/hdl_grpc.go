package plugins

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/unit-io/unitd/plugins"
	pbx "github.com/unit-io/unitd/proto"
	"github.com/unit-io/unitd/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

const (
	MaxMessageSize = 1 << 19
)

// session implements net.Conn across a gRPC stream.
//
// Methods such as
// LocalAddr, RemoteAddr, deadlines, etsess. do not work.
type session struct {
	// Stream is the stream to wrap into a Conn. This is duplex stream.
	Stream grpc.Stream

	// InMsg is the type to use for reading request data from the streaming
	// endpoint. This must be a non-nil allocated value and must NOT point to
	// the same value as OutMsg since they may be used concurrently.
	//
	// The Reset method will be called on InMsg during Reads so data you
	// set initially will be lost.
	InMsg proto.Message

	// OutMsg is the type to use for sending data to streaming endpoint. This must be
	// a non-nil allocated value and must NOT point to the same value as InMsg
	// since they may be used concurrently.
	//
	// The Reset method is never called on OutMsg so they will be sent for every request
	// unless the Encode field changes it.
	OutMsg proto.Message

	// WriteLock, if non-nil, will be locked while calling SendMsg
	// on the Stream. This can be used to prevent concurrent access to
	// SendMsg which is unsafe.
	WriteLock *sync.Mutex

	// Encode encodes messages into the Request. See Encoder for more information.
	Encode plugins.Encoder

	// Decode decodes messages from the Response into a byte slice. See
	// Decoder for more information.
	Decode plugins.Decoder

	// readOffset tracks where we've read up to if we're reading a result
	// that didn't fully fit into the target slice. See Read.
	readOffset int

	// locks to ensure that only one reader/writer are operating at a time.
	// Go documents the `net.Conn` interface as being safe for simultaneous
	// readers/writers so we need to implement locking.
	readLock, writeLock sync.Mutex
}

// Read implements io.Reader.
func (sess *session) Read(p []byte) (int, error) {
	sess.readLock.Lock()
	defer sess.readLock.Unlock()

	// Attempt to read a value only if we're not still decoding a
	// partial read value from the last result.
	if sess.readOffset == 0 {
		if err := sess.Stream.RecvMsg(sess.InMsg); err != nil {
			return 0, err
		}
	}

	// Decode into our slice
	data, err := sess.Decode(sess.InMsg, sess.readOffset, p)

	// If we have an error or we've decoded the full amount then we're done.
	// The error case is obvious. The case where we've read the full amount
	// is also a terminating condition and err == nil (we know this since its
	// checking that case) so we return the full amount and no error.
	if err != nil || len(data) <= (len(p)+sess.readOffset) {
		// Reset the read offset since we're done.
		n := len(data) - sess.readOffset
		sess.readOffset = 0

		// Reset our response value for the next read and so that we
		// don't potentially store a large response structure in memory.
		sess.InMsg.Reset()

		return n, err
	}

	// We didn't read the full amount so we need to store this for future reads
	sess.readOffset += len(p)

	return len(p), nil
}

// Write implements io.Writer.
func (sess *session) Write(p []byte) (int, error) {
	sess.writeLock.Lock()
	defer sess.writeLock.Unlock()

	total := len(p)
	for {
		// Encode our data into the request. Any error means we abort.
		n, err := sess.Encode(sess.OutMsg, p)
		if err != nil {
			return 0, err
		}

		// We lock for SendMsg if we have a lock set.
		if sess.WriteLock != nil {
			sess.WriteLock.Lock()
		}

		// Send our message. Any error we also just abort out.
		err = sess.Stream.SendMsg(sess.OutMsg)
		if sess.WriteLock != nil {
			sess.WriteLock.Unlock()
		}
		if err != nil {
			return 0, err
		}

		// If we sent the full amount of data, we're done. We respond with
		// "total" in case we sent across multiple frames.
		if n == len(p) {
			return total, nil
		}

		// We sent partial data so we continue writing the remainder
		p = p[n:]
	}
}

// Close will close the client if this is a client. If this is a server
// stream this does nothing since gRPC expects you to close the stream by
// returning from the RPC call.
//
// This calls CloseSend underneath for clients, so read the documentation
// for that to understand the semantics of this call.
func (sess *session) Close() error {
	if cs, ok := sess.Stream.(grpc.ClientStream); ok {
		// We have to acquire the write lock since the gRPC docs state:
		// "It is also not safe to call CloseSend concurrently with SendMsg."
		sess.writeLock.Lock()
		defer sess.writeLock.Unlock()
		return cs.CloseSend()
	}

	return nil
}

// LocalAddr returns nil.
func (sess *session) LocalAddr() net.Addr { return nil }

// RemoteAddr returns nil.
func (sess *session) RemoteAddr() net.Addr { return nil }

// SetDeadline is non-functional due to limitations on how gRPC works.
// You can mimic deadlines often using call options.
func (sess *session) SetDeadline(time.Time) error { return nil }

// SetReadDeadline is non-functional, see SetDeadline.
func (sess *session) SetReadDeadline(time.Time) error { return nil }

// SetWriteDeadline is non-functional, see SetDeadline.
func (sess *session) SetWriteDeadline(time.Time) error { return nil }

var _ net.Conn = (*session)(nil)

//Handler is a callback which get called when a grpc stream is established
type Handler func(c net.Conn, proto types.Proto)

type Server struct {
	Handler Handler
}

// Start implements Unitd.Start
func (s *Server) Start(ctx context.Context, info *pbx.ConnInfo) (*pbx.ConnInfo, error) {
	if info != nil {
		// Will panic if msg is not of *pbx.ConnInfo type. This is an intentional panic.
		return info, nil
	}
	return nil, nil
}

func StreamConn(
	stream grpc.Stream,
) *session {
	clientInfoFieldFunc := func(msg proto.Message) *[]byte {
		return &msg.(*pbx.ConnInfo).ClientId
	}

	return &session{
		Stream: stream,
		InMsg:  &pbx.ConnInfo{},
		OutMsg: &pbx.ConnInfo{},
		Encode: plugins.SimpleEncoder(clientInfoFieldFunc),
		Decode: plugins.SimpleDecoder(clientInfoFieldFunc),
	}
}

// Stream implements duplex Unitd.Stream
func (s *Server) Stream(stream pbx.Unitd_StreamServer) error {
	conn := StreamConn(stream)
	defer conn.Close()

	go s.Handler(conn, types.GRPC)
	<-stream.Context().Done()
	return nil
}

// Stop implements Unitd.Stop
func (s *Server) Stop(context.Context, *pbx.Empty) (*pbx.Empty, error) {
	return nil, nil
}

func (s *Server) ListenAndServe(addr string, kaEnabled bool, tlsConf *tls.Config) error {
	if addr == "" {
		return nil
	}

	lis, err := netListener(addr)
	if err != nil {
		return err
	}

	secure := ""
	var opts []grpc.ServerOption
	opts = append(opts, grpc.MaxRecvMsgSize(int(MaxMessageSize)))
	if tlsConf != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConf)))
		secure = " secure"
	}

	if kaEnabled {
		kepConfig := keepalive.EnforcementPolicy{
			MinTime:             1 * time.Second, // If a client pings more than once every second, terminate the connection
			PermitWithoutStream: true,            // Allow pings even when there are no active streams
		}
		opts = append(opts, grpc.KeepaliveEnforcementPolicy(kepConfig))

		kpConfig := keepalive.ServerParameters{
			Time:    60 * time.Second, // Ping the client if it is idle for 60 seconds to ensure the connection is still active
			Timeout: 20 * time.Second, // Wait 20 second for the ping ack before assuming the connection is dead
		}
		opts = append(opts, grpc.KeepaliveParams(kpConfig))
	}

	srv := grpc.NewServer(opts...)
	pbx.RegisterUnitdServer(srv, s)
	log.Printf("gRPC/%s%s server is registered", grpc.Version, secure)
	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Println("gRPC server failed:", err)
		}
	}()
	return nil
}

// netListener creates net.Listener for tcp and unix domains:
// if addr is is in the form "unix:/run/tinode.sock" it's a unix socket, otherwise TCP host:port.
func netListener(addr string) (net.Listener, error) {
	addrParts := strings.SplitN(addr, ":", 2)
	if len(addrParts) == 2 && addrParts[0] == "unix" {
		return net.Listen("unix", addrParts[1])
	}
	return net.Listen("tcp", addr)
}

var _ pbx.UnitdServer = (*Server)(nil)
