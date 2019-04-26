package mqtt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// varHeader reserves the bytes for a variable header.
var varHeader = []byte{0x0, 0x0, 0x0, 0x0}

//Packet is the interface all our packets will be implementing
type Packet interface {
	fmt.Stringer

	Type() uint8
	Encode() []byte
}

// MQTT message types
const (
	CONNECT = uint8(iota + 1)
	CONNACK
	PUBLISH
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PINGREQ
	PINGRESP
	DISCONNECT
)

// FixedHeader
type FixedHeader struct {
	DUP    bool
	Retain bool
	QOS    uint8
}

// unpackPacket unpacks the packet from the provided reader.
func DecodePacket(rdr io.Reader) (Packet, error) {
	hdr, sizeOf, packetType, err := unpackFixedHeader(rdr)
	if err != nil {
		return nil, err
	}

	// Check for empty packets
	switch packetType {
	case PINGREQ:
		return &Pingreq{}, nil
	case PINGRESP:
		return &Pingresp{}, nil
	case DISCONNECT:
		return &Disconnect{}, nil
	}

	buffer := make([]byte, sizeOf)
	_, err = io.ReadFull(rdr, buffer)
	if err != nil {
		return nil, err
	}

	// unpack the body
	var pkt Packet
	switch packetType {
	case CONNECT:
		pkt = unpackConnect(buffer, hdr)
	case CONNACK:
		pkt = unpackConnack(buffer, hdr)
	case PUBLISH:
		pkt = unpackPublish(buffer, hdr)
	case PUBACK:
		pkt = unpackPuback(buffer, hdr)
	case PUBREC:
		pkt = unpackPubrec(buffer, hdr)
	case PUBREL:
		pkt = unpackPubrel(buffer, hdr)
	case PUBCOMP:
		pkt = unpackPubcomp(buffer, hdr)
	case SUBSCRIBE:
		pkt = unpackSubscribe(buffer, hdr)
	case SUBACK:
		pkt = unpackSuback(buffer, hdr)
	case UNSUBSCRIBE:
		pkt = unpackUnsubscribe(buffer, hdr)
	case UNSUBACK:
		pkt = unpackUnsuback(buffer, hdr)
	default:
		return nil, fmt.Errorf("Invalid zero-length packet with type %d", packetType)
	}

	return pkt, nil
}

// encodeParts sews the whole packet together
func encodeParts(pktType uint8, length int, h *FixedHeader) bytes.Buffer {
	var buf bytes.Buffer
	var firstByte byte
	firstByte |= pktType << 4
	if h != nil {
		firstByte |= boolToUInt8(h.DUP) << 3
		firstByte |= h.QOS << 1
		firstByte |= boolToUInt8(h.Retain)
	}
	buf.WriteByte(firstByte)
	buf.Write(encodeLength(length))
	return buf
}

// unpackFixedHeader unpacks the header
func unpackFixedHeader(rdr io.Reader) (hdr *FixedHeader, length uint32, packetType uint8, err error) {
	b := make([]byte, 1)
	if _, err = io.ReadFull(rdr, b); err != nil {
		return nil, 0, 0, err
	}

	controlByte := b[0]
	packetType = (controlByte & 0xf0) >> 4

	// Set the header depending on the packet type
	switch packetType {
	case PUBLISH, SUBSCRIBE, UNSUBSCRIBE, PUBREL:
		DUP := controlByte&0x08 > 0
		QOS := controlByte & 0x06 >> 1
		retain := controlByte&0x01 > 0

		hdr = &FixedHeader{
			DUP:    DUP,
			QOS:    QOS,
			Retain: retain,
		}
	}

	return hdr, decodeLength(rdr), packetType, nil
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

func decodeLength(r io.Reader) uint32 {
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
	return rLength
}
