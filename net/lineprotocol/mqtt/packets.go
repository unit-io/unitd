package mqtt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	lp "github.com/unit-io/unitd/net/lineprotocol"
)

// varHeader reserves the bytes for a variable header.
var varHeader = []byte{0x0, 0x0, 0x0, 0x0}

type Packet lp.Packet

type FixedHeader lp.FixedHeader

// ReadPacket unpacks the packet from the provided reader.
func ReadPacket(r io.Reader) (Packet, error) {
	var fh FixedHeader
	if err := fh.unpack(r); err != nil {
		return nil, err
	}

	// Check for empty packets
	switch fh.MessageType {
	case lp.PINGREQ:
		return &Pingreq{}, nil
	case lp.PINGRESP:
		return &Pingresp{}, nil
	case lp.DISCONNECT:
		return &Disconnect{}, nil
	}

	buffer := make([]byte, fh.RemainingLength)
	if _, err := io.ReadFull(r, buffer); err != nil {
		return nil, err
	}

	// unpack the body
	var pkt Packet
	switch fh.MessageType {
	case lp.CONNECT:
		pkt = unpackConnect(buffer, fh)
	case lp.CONNACK:
		pkt = unpackConnack(buffer, fh)
	case lp.PUBLISH:
		pkt = unpackPublish(buffer, fh)
	case lp.PUBACK:
		pkt = unpackPuback(buffer, fh)
	case lp.PUBREC:
		pkt = unpackPubrec(buffer, fh)
	case lp.PUBREL:
		pkt = unpackPubrel(buffer, fh)
	case lp.PUBCOMP:
		pkt = unpackPubcomp(buffer, fh)
	case lp.SUBSCRIBE:
		pkt = unpackSubscribe(buffer, fh)
	case lp.SUBACK:
		pkt = unpackSuback(buffer, fh)
	case lp.UNSUBSCRIBE:
		pkt = unpackUnsubscribe(buffer, fh)
	case lp.UNSUBACK:
		pkt = unpackUnsuback(buffer, fh)
	default:
		return nil, fmt.Errorf("Invalid zero-length packet with type %d", fh.MessageType)
	}

	return pkt, nil
}

func (fh *FixedHeader) pack(h *lp.FixedHeader) bytes.Buffer {
	var buf bytes.Buffer
	var firstByte byte
	firstByte |= fh.MessageType << 4
	if h != nil {
		firstByte |= boolToUInt8(h.DUP) << 3
		firstByte |= h.QOS << 1
		firstByte |= boolToUInt8(h.Retain)
	}
	buf.WriteByte(firstByte)
	buf.Write(encodeLength(fh.RemainingLength))
	return buf
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
		fh.DUP = controlByte&0x08 > 0
		fh.QOS = controlByte & 0x06 >> 1
		fh.Retain = controlByte&0x01 > 0
	}

	fh.RemainingLength = decodeLength(r)
	return nil
}

// -------------------------------------------------------------
func writeString(buf *bytes.Buffer, s []byte) {
	strlen := uint16(len(s))
	writeUint16(buf, strlen)
	buf.Write(s)
}

func writeUint16(buf *bytes.Buffer, tupac uint16) {
	buf.WriteByte(byte((tupac & 0xff00) >> 8))
	buf.WriteByte(byte(tupac & 0x00ff))
}

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

func boolToByte(b bool) byte {
	switch b {
	case true:
		return 1
	default:
		return 0
	}
}

func decodeByte(b io.Reader) byte {
	num := make([]byte, 1)
	b.Read(num)
	return num[0]
}

func decodeUint16(b io.Reader) uint16 {
	num := make([]byte, 2)
	b.Read(num)
	return binary.BigEndian.Uint16(num)
}

func encodeUint16(num uint16) []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, num)
	return bytes
}

func encodeString(field string) []byte {
	fieldLength := make([]byte, 2)
	binary.BigEndian.PutUint16(fieldLength, uint16(len(field)))
	return append(fieldLength, []byte(field)...)
}

func decodeString(b io.Reader) string {
	fieldLength := decodeUint16(b)
	field := make([]byte, fieldLength)
	b.Read(field)
	return string(field)
}

func decodeBytes(b io.Reader) []byte {
	fieldLength := decodeUint16(b)
	field := make([]byte, fieldLength)
	b.Read(field)
	return field
}

func encodeBytes(field []byte) []byte {
	fieldLength := make([]byte, 2)
	binary.BigEndian.PutUint16(fieldLength, uint16(len(field)))
	return append(fieldLength, field...)
}

func varHeaderLength(length int) int {
	var varBytes int
	for {
		length /= 128
		if length == 0 {
			break
		}
		varBytes++
	}
	return varBytes
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
