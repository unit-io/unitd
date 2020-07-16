package mqtt

import (
	"bytes"
	"errors"

	lp "github.com/unit-io/unitd/lineprotocol"
)

func encodePublish(p lp.Publish) (bytes.Buffer, error) {
	var msg bytes.Buffer
	var length int
	length = 2 + len(p.Topic) + len(p.Payload)
	if p.FixedHeader.Qos > 0 {
		length += 2
	}
	if length > lp.MaxMessageSize {
		return msg, errors.New("Message too large")
	}

	msg.Write(encodeBytes(p.Topic))
	if p.FixedHeader.Qos > 0 {
		msg.Write(encodeUint16(p.MessageID))
	}
	msg.Write(p.Payload)
	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.PUBLISH, RemainingLength: length}
	packet := fh.pack(&p.FixedHeader)
	packet.Write(msg.Bytes())
	return packet, nil
}

func encodePuback(p lp.Puback) (bytes.Buffer, error) {
	fh := FixedHeader{MessageType: lp.PUBACK, RemainingLength: 2}
	packet := fh.pack(nil)
	_, err := packet.Write(encodeUint16(p.MessageID))
	return packet, err
}

func encodePubrec(p lp.Pubrec) (bytes.Buffer, error) {
	fh := FixedHeader{MessageType: lp.PUBREC, RemainingLength: 2}
	packet := fh.pack(&p.FixedHeader)
	_, err := packet.Write(encodeUint16(p.MessageID))
	return packet, err
}

func encodePubrel(p lp.Pubrel) (bytes.Buffer, error) {
	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.PUBREL, RemainingLength: 2}
	packet := fh.pack(&p.FixedHeader)
	_, err := packet.Write(encodeUint16(p.MessageID))
	return packet, err
}

func encodePubcomp(p lp.Pubcomp) (bytes.Buffer, error) {
	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.PUBCOMP, RemainingLength: 2}
	packet := fh.pack(nil)
	_, err := packet.Write(encodeUint16(p.MessageID))
	return packet, err
}

func unpackPublish(data []byte, fh FixedHeader) lp.Packet {
	bookmark := uint32(0)
	topic := readString(data, &bookmark)
	var msgID uint16
	if fh.Qos > 0 {
		msgID = readUint16(data, &bookmark)
	}

	return &lp.Publish{
		FixedHeader: lp.FixedHeader(fh),
		Topic:       topic,
		Payload:     data[bookmark:],
		MessageID:   msgID,
	}
}

func unpackPuback(data []byte, fh FixedHeader) lp.Packet {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &lp.Puback{
		MessageID: msgID,
	}
}

func unpackPubrec(data []byte, fh FixedHeader) lp.Packet {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &lp.Pubrec{
		FixedHeader: lp.FixedHeader(fh),
		MessageID:   msgID,
	}
}

func unpackPubrel(data []byte, fh FixedHeader) lp.Packet {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &lp.Pubrel{
		FixedHeader: lp.FixedHeader(fh),
		MessageID:   msgID,
	}
}

func unpackPubcomp(data []byte, fh FixedHeader) lp.Packet {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &lp.Pubcomp{
		MessageID: msgID,
	}
}
