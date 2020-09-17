package mqtt

import (
	"bytes"

	lp "github.com/unit-io/unitd/lineprotocol"
)

func encodeSubscribe(s lp.Subscribe) (bytes.Buffer, error) {
	var msg bytes.Buffer

	//msg.Write(reserveForHeader)
	msg.Write(encodeUint16(s.MessageID))
	for _, sub := range s.Subscriptions {
		msg.Write(encodeBytes(sub.Topic))
		msg.WriteByte(byte(sub.Qos))
	}

	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.SUBSCRIBE, RemainingLength: msg.Len()}
	packet := fh.pack(&s.FixedHeader)
	_, err := packet.Write(msg.Bytes())
	return packet, err
}

func encodeSuback(s lp.Suback) (bytes.Buffer, error) {
	var msg bytes.Buffer

	//msg.Write(reserveForHeader)
	msg.Write(encodeUint16(s.MessageID))
	for _, q := range s.Qos {
		msg.WriteByte(byte(q))
	}

	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.SUBACK, RemainingLength: msg.Len()}
	packet := fh.pack(nil)
	_, err := packet.Write(msg.Bytes())
	return packet, err
}

func encodeUnsubscribe(u lp.Unsubscribe) (bytes.Buffer, error) {
	var msg bytes.Buffer

	//msg.Write(reserveForHeader)
	msg.Write(encodeUint16(u.MessageID))
	for _, sub := range u.Subscriptions {
		msg.Write(encodeBytes(sub.Topic))
	}

	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.UNSUBSCRIBE, RemainingLength: msg.Len()}
	packet := fh.pack(&u.FixedHeader)
	_, err := packet.Write(msg.Bytes())
	return packet, err
}

func encodeUnsuback(u lp.Unsuback) (bytes.Buffer, error) {
	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.UNSUBACK, RemainingLength: 2}
	packet := fh.pack(nil)
	_, err := packet.Write(encodeUint16(u.MessageID))
	return packet, err
}

func unpackSubscribe(data []byte, fh FixedHeader) lp.Packet {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	var topics []lp.TopicQOSTuple
	maxlen := uint32(len(data))
	for bookmark < maxlen {
		var t lp.TopicQOSTuple
		t.Topic = readString(data, &bookmark)
		qos := data[bookmark]
		bookmark++
		t.Qos = uint8(qos)
		topics = append(topics, t)
	}
	return &lp.Subscribe{
		FixedHeader:   lp.FixedHeader(fh),
		MessageID:     msgID,
		Subscriptions: topics,
	}
}

func unpackSuback(data []byte, fh FixedHeader) lp.Packet {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	var qoses []uint8
	maxlen := uint32(len(data))
	for bookmark < maxlen {
		qos := data[bookmark]
		bookmark++
		qoses = append(qoses, qos)
	}
	return &lp.Suback{
		MessageID: msgID,
		Qos:       qoses,
	}
}

func unpackUnsubscribe(data []byte, fh FixedHeader) lp.Packet {
	bookmark := uint32(0)
	var topics []lp.TopicQOSTuple
	msgID := readUint16(data, &bookmark)
	maxlen := uint32(len(data))
	for bookmark < maxlen {
		var t lp.TopicQOSTuple
		t.Topic = readString(data, &bookmark)
		topics = append(topics, t)
	}
	return &lp.Unsubscribe{
		FixedHeader:   lp.FixedHeader(fh),
		MessageID:     msgID,
		Subscriptions: topics,
	}
}

func unpackUnsuback(data []byte, fh FixedHeader) lp.Packet {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &lp.Unsuback{
		MessageID: msgID,
	}
}
