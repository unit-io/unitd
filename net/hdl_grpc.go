package net

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	plugins "github.com/unit-io/unitd/plugins/grpc"
	pbx "github.com/unit-io/unitd/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

const (
	MaxMessageSize = 1 << 19
)

//Handler is a callback which get called when a grpc stream is established
type StreamHandler func(c net.Conn, proto Proto)

// type grpcServer struct {
// }

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
) *plugins.Conn {
	packetFunc := func(msg proto.Message) *[]byte {
		return &msg.(*pbx.Packet).Data
	}
	return &plugins.Conn{
		Stream: stream,
		InMsg:  &pbx.Packet{},
		OutMsg: &pbx.Packet{},
		Encode: plugins.Encode(packetFunc),
		Decode: plugins.Decode(packetFunc),
	}
}

// Stream implements duplex Unitd.Stream
func (s *Server) Stream(stream pbx.Unitd_StreamServer) error {
	conn := StreamConn(stream)
	defer conn.Close()

	go s.StreamHandler(conn, GRPC)
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
