package main

import (
	"context"
	"io"
	"log"
	"time"

	pbx "github.com/unit-io/unitd/proto"
	"google.golang.org/grpc"
)

func main() {
	// keygen request
	// generateKey()

	// Connect to the server
	conn, err := grpc.Dial(
		"localhost:6061",
		grpc.WithBlock(),
		grpc.WithInsecure())
	if err != nil {
		log.Fatalf("err: %s", err)
	}
	defer func() {
		conn.Close()
	}()

	// Connect for streaming
	r, err := pbx.NewUnitdClient(conn).Stream(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	go func(retry int) {
		i := 0
		for range time.Tick(1 * time.Second) {
			// var pkt pbx.InMsg
			// pkt.Message = &pbx.InMsg_Info{Info: &pbx.ConnInfo{ClientId: []byte("UCBFDONCNJLaKMCAIeJBaOVfbAXUZHNPLDKKLDKLHZHKYIZLCDPQ")}}
			r.SendMsg(&pbx.ConnInfo{ClientId: []byte("UCBFDONCNJLaKMCAIeJBaOVfbAXUZHNPLDKKLDKLHZHKYIZLCDPQ")})
			if i >= retry {
				return
			}
			i++
		}
	}(10)

	for {
		select {
		case <-r.Context().Done():
			return
		default:
			// var in pbx.ConnInfo
			in, err := r.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Println("grpc: recv", err)
				log.Fatal(err)
			}
			log.Println("grpc in:", in.String())
		}
	}
}
