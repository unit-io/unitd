package grpc

import (
	"bytes"
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"
	lp "github.com/unit-io/unitd/net"
	pbx "github.com/unit-io/unitd/proto"
)

//Packet is the interface all our packets in the line protocol will be implementing
type Packet lp.Packet

type FixedHeader pbx.FixedHeader

// ReadPacket unpacks the packet from the provided reader.
func ReadPacket(r io.Reader) (Packet, error) {
	var fh FixedHeader
	fh.unpack(r)

	// Check for empty packets
	switch fh.MessageType {
	case pbx.MessageType_PINGREQ:
		return &Pingreq{}, nil
	case pbx.MessageType_PINGRESP:
		return &Pingresp{}, nil
	case pbx.MessageType_DISCONNECT:
		return &Disconnect{}, nil
	}

	buf := make([]byte, fh.RemainingLength)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	// unpack the body
	var pkt Packet
	switch fh.MessageType {
	case pbx.MessageType_CONNECT:
		pkt = unpackConnect(buf)
	case pbx.MessageType_CONNACK:
		pkt = unpackConnack(buf)
	case pbx.MessageType_PUBLISH:
		pkt = unpackPublish(buf)
	case pbx.MessageType_PUBACK:
		pkt = unpackPuback(buf)
	case pbx.MessageType_PUBREC:
		pkt = unpackPubrec(buf)
	case pbx.MessageType_PUBREL:
		pkt = unpackPubrel(buf)
	case pbx.MessageType_PUBCOMP:
		pkt = unpackPubcomp(buf)
	case pbx.MessageType_SUBSCRIBE:
		pkt = unpackSubscribe(buf)
	case pbx.MessageType_SUBACK:
		pkt = unpackSuback(buf)
	case pbx.MessageType_UNSUBSCRIBE:
		pkt = unpackUnsubscribe(buf)
	case pbx.MessageType_UNSUBACK:
		pkt = unpackUnsuback(buf)
	default:
		return nil, fmt.Errorf("Invalid zero-length packet with type %d", fh.MessageType)
	}

	return pkt, nil
}

func (fh *FixedHeader) pack() bytes.Buffer {
	var buf bytes.Buffer
	ph := pbx.FixedHeader(*fh)
	h, err := proto.Marshal(&ph)
	if err != nil {
		return buf
	}
	size := encodeLength(len(h))
	buf.Write(size)
	buf.Write(h)
	return buf
}

func (fh *FixedHeader) unpack(r io.Reader) error {
	fhSize, err := decodeLength(r)
	if err != nil {
		return err
	}

	// read FixedHeader
	buf := make([]byte, fhSize)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return err
	}

	var h pbx.FixedHeader
	proto.Unmarshal(buf, &h)

	*fh = FixedHeader(h)
	return nil
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
