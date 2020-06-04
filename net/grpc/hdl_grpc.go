package grpc

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	pbx "github.com/unit-io/unitd/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

const (
	MaxMessageSize = 1 << 19
)

// session implements net.Conn across a gRPC stream. You must populate many
// of the exported structs on this field so please read the documentation.
//
// There are a number of limitations to this implementation, typically due
// limitations of visibility given by the gRPC stream. Methods such as
// LocalAddr, RemoteAddr, deadlines, etsess. do not work.
//
// As documented on net.Conn, it is safe for concurrent read/write.
type session struct {
	// Stream is the stream to wrap into a Conn. This can be either a client
	// or server stream and we will perform correctly.
	Stream grpc.Stream

	// Request is the type to use for sending request data to the streaming
	// endpoint. This must be a non-nil allocated value and must NOT point to
	// the same value as Response since they may be used concurrently.
	//
	// The Reset method is never called on Request so you may set some
	// fields on the request type and they will be sent for every request
	// unless the Encode field changes it.
	Request proto.Message

	// Response is the type to use for reading response data. This must be
	// a non-nil allocated value and must NOT point to the same value as Request
	// since they may be used concurrently.
	//
	// The Reset method will be called on Response during Reads so data you
	// set initially will be lost.
	Response proto.Message

	// ResponseLock, if non-nil, will be locked while calling SendMsg
	// on the Stream. This can be used to prevent concurrent access to
	// SendMsg which is unsafe.
	ResponseLock *sync.Mutex

	// Encode encodes messages into the Request. See Encoder for more information.
	Encode Encoder

	// Decode decodes messages from the Response into a byte slice. See
	// Decoder for more information.
	Decode Decoder

	reader  io.Reader
	closing chan bool
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
		if err := sess.Stream.RecvMsg(sess.Response); err != nil {
			return 0, err
		}
	}

	// Decode into our slice
	data, err := sess.Decode(sess.Response, sess.readOffset, p)

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
		sess.Response.Reset()

		return n, err
	}

	// We didn't read the full amount so we need to store this for future reads
	sess.readOffset += len(p)

	// Read from the reader
	n, err := sess.reader.Read(p)
	if err != nil {
		if err == io.EOF {
			sess.reader = nil
			err = nil
		}
	}

	return n, nil
}

// Write implements io.Writer.
func (sess *session) Write(p []byte) (int, error) {
	sess.writeLock.Lock()
	defer sess.writeLock.Unlock()

	total := len(p)
	for {
		// Encode our data into the request. Any error means we abort.
		n, err := sess.Encode(sess.Request, p)
		if err != nil {
			return 0, err
		}

		// We lock for SendMsg if we have a lock set.
		if sess.ResponseLock != nil {
			sess.ResponseLock.Lock()
		}

		// Send our message. Any error we also just abort out.
		err = sess.Stream.SendMsg(sess.Request)
		if sess.ResponseLock != nil {
			sess.ResponseLock.Unlock()
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

type grpcServer struct {
}

// Start implements Unitd.Start
func (s *grpcServer) Start(ctx context.Context, info *pbx.ConnInfo) (*pbx.ConnInfo, error) {
	if info != nil {
		// Will panic if msg is not of *pbx.ConnInfo type. This is an intentional panic.
		return info, nil
	}
	return nil, nil
}

// Stream implements duplex Unitd.Stream
func (s *grpcServer) Stream(stream pbx.Unitd_StreamServer) error {
	if stream != nil {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Println("ConnInfo: recv", err)
			return err
		}

		entry := in.GetEntry()
		var info interface{}
		info = &pbx.Info{Topic: entry.Topic, Seq: 1}

		// Will panic if msg is not of *pbx.ConnInfo type. This is an intentional panic.
		return stream.Send(info.(*pbx.ServerMsg))
	}
	return nil
}

// Stop implements Unitd.Stop
func (s *grpcServer) Stop(context.Context, *pbx.Empty) (*pbx.Empty, error) {
	return nil, nil
}
func New(addr string, kaEnabled bool, tlsConf *tls.Config) *grpc.Server {
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

	svr := grpc.NewServer(opts...)
	pbx.RegisterUnitdServer(svr, &grpcServer{})
	log.Printf("gRPC/%s%s server is registered", grpc.Version, secure)

	return svr
}
