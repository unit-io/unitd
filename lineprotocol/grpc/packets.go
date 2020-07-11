package grpc

import (
	"bytes"
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"
	lp "github.com/unit-io/unitd/lineprotocol"
	pbx "github.com/unit-io/unitd/proto"
)

type FixedHeader pbx.FixedHeader

type LineProto struct {
}

// ReadPacket unpacks the packet from the provided reader.
func (p *LineProto) ReadPacket(r io.Reader) (lp.Packet, error) {
	var fh FixedHeader
	fh.unpack(r)

	// Check for empty packets
	switch fh.MessageType {
	case pbx.MessageType_PINGREQ:
		return &lp.Pingreq{}, nil
	case pbx.MessageType_PINGRESP:
		return &lp.Pingresp{}, nil
	case pbx.MessageType_DISCONNECT:
		return &lp.Disconnect{}, nil
	}

	msg := make([]byte, fh.RemainingLength)
	_, err := io.ReadFull(r, msg)
	if err != nil {
		return nil, err
	}

	// unpack the body
	var pkt lp.Packet
	switch fh.MessageType {
	case pbx.MessageType_CONNECT:
		pkt = unpackConnect(msg)
	case pbx.MessageType_CONNACK:
		pkt = unpackConnack(msg)
	case pbx.MessageType_PUBLISH:
		pkt = unpackPublish(msg)
	case pbx.MessageType_PUBACK:
		pkt = unpackPuback(msg)
	case pbx.MessageType_PUBREC:
		pkt = unpackPubrec(msg)
	case pbx.MessageType_PUBREL:
		pkt = unpackPubrel(msg)
	case pbx.MessageType_PUBCOMP:
		pkt = unpackPubcomp(msg)
	case pbx.MessageType_SUBSCRIBE:
		pkt = unpackSubscribe(msg)
	case pbx.MessageType_SUBACK:
		pkt = unpackSuback(msg)
	case pbx.MessageType_UNSUBSCRIBE:
		pkt = unpackUnsubscribe(msg)
	case pbx.MessageType_UNSUBACK:
		pkt = unpackUnsuback(msg)
	default:
		return nil, fmt.Errorf("Invalid zero-length packet with type %d", fh.MessageType)
	}

	return pkt, nil
}

// Encode encodes the message into binary data
func (p *LineProto) Encode(pkt lp.Packet) (bytes.Buffer, error) {
	switch pkt.Type() {
	case lp.PINGREQ:
		return encodePingreq(*pkt.(*lp.Pingreq))
	case lp.CONNACK:
		return encodeConnack(*pkt.(*lp.Connack))
	case lp.SUBACK:
		return encodeSuback(*pkt.(*lp.Suback))
	case lp.UNSUBACK:
		return encodeUnsuback(*pkt.(*lp.Unsuback))
	case lp.PUBLISH:
		return encodePublish(*pkt.(*lp.Publish))
	case lp.PUBACK:
		return encodePuback(*pkt.(*lp.Puback))
	case lp.PUBREC:
		return encodePubrec(*pkt.(*lp.Pubrec))
	case lp.PUBREL:
		return encodePubrel(*pkt.(*lp.Pubrel))
	case lp.PUBCOMP:
		return encodePubcomp(*pkt.(*lp.Pubcomp))
	}
	return bytes.Buffer{}, nil
}

func (fh *FixedHeader) pack() bytes.Buffer {
	var head bytes.Buffer
	ph := pbx.FixedHeader(*fh)
	h, err := proto.Marshal(&ph)
	if err != nil {
		return head
	}
	size := encodeLength(len(h))
	head.Write(size)
	head.Write(h)
	return head
}

func (fh *FixedHeader) unpack(r io.Reader) error {
	fhSize, err := decodeLength(r)
	if err != nil {
		return err
	}

	// read FixedHeader
	head := make([]byte, fhSize)
	_, err = io.ReadFull(r, head)
	if err != nil {
		return err
	}

	var h pbx.FixedHeader
	proto.Unmarshal(head, &h)

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
