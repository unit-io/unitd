package mqtt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	lp "github.com/unit-io/unitd/lineprotocol"
)

type FixedHeader lp.FixedHeader

type LineProto struct {
}

// ReadPacket unpacks the packet from the provided reader.
func (p *LineProto) ReadPacket(r io.Reader) (lp.Packet, error) {
	var fh FixedHeader
	if err := fh.unpack(r); err != nil {
		return nil, err
	}

	// Check for empty packets
	switch fh.MessageType {
	case lp.PINGREQ:
		return &lp.Pingreq{}, nil
	case lp.PINGRESP:
		return &lp.Pingresp{}, nil
	case lp.DISCONNECT:
		return &lp.Disconnect{}, nil
	}

	msg := make([]byte, fh.RemainingLength)
	if _, err := io.ReadFull(r, msg); err != nil {
		return nil, err
	}

	// unpack the body
	var pkt lp.Packet
	switch fh.MessageType {
	case lp.CONNECT:
		pkt = unpackConnect(msg, fh)
	case lp.CONNACK:
		pkt = unpackConnack(msg, fh)
	case lp.PUBLISH:
		pkt = unpackPublish(msg, fh)
	case lp.PUBACK:
		pkt = unpackPuback(msg, fh)
	case lp.PUBREC:
		pkt = unpackPubrec(msg, fh)
	case lp.PUBREL:
		pkt = unpackPubrel(msg, fh)
	case lp.PUBCOMP:
		pkt = unpackPubcomp(msg, fh)
	case lp.SUBSCRIBE:
		pkt = unpackSubscribe(msg, fh)
	case lp.SUBACK:
		pkt = unpackSuback(msg, fh)
	case lp.UNSUBSCRIBE:
		pkt = unpackUnsubscribe(msg, fh)
	case lp.UNSUBACK:
		pkt = unpackUnsuback(msg, fh)
	default:
		return nil, fmt.Errorf("Invalid zero-length packet with type %d", fh.MessageType)
	}

	return pkt, nil
}

func (p *LineProto) Encode(pkt lp.Packet) (bytes.Buffer, error) {
	switch pkt.Type() {
	case lp.PINGREQ:
		return encodePingreq(*pkt.(*lp.Pingreq))
	case lp.PINGRESP:
		return encodePingresp(*pkt.(*lp.Pingresp))
	case lp.CONNECT:
		return encodeConnect(*pkt.(*lp.Connect))
	case lp.CONNACK:
		return encodeConnack(*pkt.(*lp.Connack))
	case lp.DISCONNECT:
		return encodeDisconnect(*pkt.(*lp.Disconnect))
	case lp.SUBSCRIBE:
		return encodeSubscribe(*pkt.(*lp.Subscribe))
	case lp.SUBACK:
		return encodeSuback(*pkt.(*lp.Suback))
	case lp.UNSUBSCRIBE:
		return encodeUnsubscribe(*pkt.(*lp.Unsubscribe))
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

func (fh *FixedHeader) pack(h *lp.FixedHeader) bytes.Buffer {
	var head bytes.Buffer
	var firstByte byte
	firstByte |= fh.MessageType << 4
	if h != nil {
		firstByte |= boolToUInt8(h.Dup) << 3
		firstByte |= h.Qos << 1
		firstByte |= boolToUInt8(h.Retain)
	}
	head.WriteByte(firstByte)
	head.Write(encodeLength(fh.RemainingLength))
	return head
}

func (fh *FixedHeader) unpack(r io.Reader) error {
	b := make([]byte, 1)
	if _, err := io.ReadFull(r, b); err != nil {
		return err
	}
	controlByte := b[0]
	fh.MessageType = (controlByte & 0xf0) >> 4
	// Set the header depending on the packet type
	switch fh.MessageType {
	case lp.PUBLISH, lp.SUBSCRIBE, lp.UNSUBSCRIBE, lp.PUBREL:
		fh.Dup = controlByte&0x08 > 0
		fh.Qos = controlByte & 0x06 >> 1
		fh.Retain = controlByte&0x01 > 0
	}

	fh.RemainingLength = decodeLength(r)
	return nil
}

// -------------------------------------------------------------
func readString(b []byte, startsAt *uint32) []byte {
	l := readUint16(b, startsAt)
	v := b[*startsAt : uint32(l)+*startsAt]
	*startsAt += uint32(l)
	return v
}

func readUint16(b []byte, startsAt *uint32) uint16 {
	fst := uint16(b[*startsAt])
	*startsAt++
	snd := uint16(b[*startsAt])
	*startsAt++
	return (fst << 8) + snd
}

func boolToUInt8(v bool) uint8 {
	if v {
		return 0x1
	}

	return 0x0
}

func encodeUint16(num uint16) []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, num)
	return bytes
}

func encodeBytes(field []byte) []byte {
	fieldLength := make([]byte, 2)
	binary.BigEndian.PutUint16(fieldLength, uint16(len(field)))
	return append(fieldLength, field...)
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

func decodeLength(r io.Reader) int {
	var rLength uint32
	var multiplier uint32 = 0
	b := make([]byte, 1)
	for {
		io.ReadFull(r, b)
		digit := b[0]
		rLength |= uint32(digit&127) << multiplier
		if (digit & 128) == 0 {
			break
		}
		multiplier += 7
	}
	return int(rLength)
}
