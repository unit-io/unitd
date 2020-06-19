package main

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log"
	"time"

	"github.com/golang/protobuf/proto"
	plugins "github.com/unit-io/unitd/plugins/grpc"
	pbx "github.com/unit-io/unitd/proto"
	"google.golang.org/grpc"
)

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

func encodeLength(length int) []byte {
	var encLength []byte
	for {
		digit := byte(length % 128)
		length /= 128
		if length > 0 {
			digit |= 0x80
		}
		encLength = append(encLength, digit)
		if length == 0 {
			break
		}
	}
	return encLength
}

func decodeLength(r io.Reader) (int, error) {
	var rLength uint32
	var multiplier uint32
	b := make([]byte, 1)
	for multiplier < 27 { //fix: Infinite '(digit & 128) == 1' will cause the dead loop
		_, err := io.ReadFull(r, b)
		if err != nil {
			return 0, err
		}

		digit := b[0]
		rLength |= uint32(digit&127) << multiplier
		if (digit & 128) == 0 {
			break
		}
		multiplier += 7
	}
	return int(rLength), nil
}

// unpackFixedHeader unpacks the header
func unpackFixedHeader(rdr io.Reader) (hdr pbx.FixedHeader, err error) {
	hdrSize, _ := decodeLength(rdr)

	// read FixedHeader
	buffer := make([]byte, hdrSize)
	_, err = io.ReadFull(rdr, buffer)
	if err != nil {
		return pbx.FixedHeader{}, err
	}

	var fh pbx.FixedHeader
	proto.Unmarshal(buffer, &fh)

	return fh, nil
}

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
	stream, err := pbx.NewUnitdClient(conn).Stream(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	clientConn := StreamConn(stream)

	reader := bufio.NewReaderSize(clientConn, 65536)

	var buffer bytes.Buffer
	c := pbx.Conn{ClientID: []byte("UCBFDONCNJLaKMCAIeJBaOVfbAXUZHNPLDKKLDKLHZHKYIZLCDPQ")}
	pkt, err := proto.Marshal(&c)
	fh := pbx.FixedHeader{MessageType: pbx.MessageType_CONNECT, RemainingLength: uint32(len(pkt))}
	header, err := proto.Marshal(&fh)
	size := encodeLength(len(header))
	buffer.Write(size)
	buffer.Write(header)
	buffer.Write(pkt)
	if err != nil {
		log.Println("grpc: recv", err)
		log.Fatal(err)
	}
	go func(retry int) {
		i := 0
		for range time.Tick(1 * time.Second) {
			clientConn.Write(buffer.Bytes())
			if i >= retry {
				return
			}
			i++
		}
	}(10)

	for {
		select {
		case <-stream.Context().Done():
			return
		default:
			fh, err := unpackFixedHeader(reader)
			if err != nil {
				log.Println("grpc: recv", err)
				log.Fatal(err)
			}

			// Check for empty packets
			switch fh.MessageType {
			case pbx.MessageType_PINGREQ:
				return
			case pbx.MessageType_PINGRESP:
				return
			case pbx.MessageType_DISCONNECT:
				return
			}

			buffer := make([]byte, fh.RemainingLength)
			n, err := clientConn.Read(buffer)
			if err != nil {
				log.Println("grpc: recv", n, err)
				log.Fatal(err)
			}

			switch fh.MessageType {
			case pbx.MessageType_CONNECT:
				var conn pbx.Conn
				proto.Unmarshal(buffer, &conn)
				log.Println("grpc in:", conn.ClientID)
			case pbx.MessageType_CONNACK:
				var connack pbx.Connack
				proto.Unmarshal(buffer, &connack)
				log.Println("grpc in:", connack.ReturnCode)
			case pbx.MessageType_PUBLISH:
			case pbx.MessageType_PUBACK:
			case pbx.MessageType_PUBREC:
			case pbx.MessageType_PUBREL:
			case pbx.MessageType_PUBCOMP:
			case pbx.MessageType_SUBSCRIBE:
			case pbx.MessageType_SUBACK:
			case pbx.MessageType_UNSUBSCRIBE:
			case pbx.MessageType_UNSUBACK:
			default:
				log.Println("Invalid zero-length packet with type ", fh.MessageType)
				return
			}
		}
	}
}
