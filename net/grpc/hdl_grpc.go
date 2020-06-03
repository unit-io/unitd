package grpc

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"sync"
	"time"

	pbx "github.com/unit-io/unitd/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

const (
	MaxMessageSize = 1 << 19
)

// session represents a grpc connection.
type session struct {
	sync.Mutex
	// gRPC handle. Set only for gRPC clients.
	grpcnode pbx.Unitd_StreamServer
	reader   io.Reader
	closing  chan bool
}

type grpcNodeServer struct {
}

func (sess *session) closeGrpc() {
	sess.Lock()
	defer sess.Unlock()
	sess.grpcnode = nil
}

func (*grpcNodeServer) Stream(stream *pbx.Unitd_StreamServer) (*pbx.ServerMsg, error) {
	sess := newConn(stream)

	go sess.Read()
	go sess.Write()

	return nil, nil
}

// newConn creates a new transport from grpc.
func newConn(conn interface{}) net.Conn {
	sess := &session{
		grpcnode: conn,
		closing:  make(chan bool),
	}

	return sess
}

func (sess *session) Read(b []byte) (n int, err error) {
	if sess.reader == nil {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil, nil
		}
		if err != nil {
			log.Println("grpc: recv", err)
			return
		}
		b := pbCliDeserialize(in)

		sess.Lock()
		if sess.grpcnode == nil {
			sess.Unlock()
			break
		}
		sess.Unlock()
	}
	// Read from the reader
	n, err = sess.reader.Read(b)
	if err != nil {
		if err == io.EOF {
			sess.reader = nil
			err = nil
		}
	}
	return
}

func (sess *session) Write(b []byte) (n int, err error) {
	// Serialize write to avoid concurrent write
	sess.Lock()
	defer func() {
		sess.Unlock()
	}()

	sess.grpcnode.Send(b.(*pbx.ServerMsg))
	return
}

// Close terminates the connection.
func (sess *session) Close() error {
	return sess.closeGrpc()
}

func serveGrpc(addr string, kaEnabled bool, tlsConf *tls.Config) (*grpc.Server, error) {
	if addr == "" {
		return nil, nil
	}

	lis, err := netListener(addr)
	if err != nil {
		return nil, err
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
	pbx.RegisterNodeServer(srv, &grpcNodeServer{})
	log.Printf("gRPC/%s%s server is registered at [%s]", grpc.Version, secure, addr)

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Println("gRPC server failed:", err)
		}
	}()

	return srv, nil
}
