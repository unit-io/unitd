package mqtt

import (
	"bytes"
	"io"

	lp "github.com/unit-io/unitd/net/lineprotocol"
)

type (
	Subscribe   lp.Subscribe
	Suback      lp.Suback
	Unsubscribe lp.Unsubscribe
	Unsuback    lp.Unsuback
)

// WriteTo writes the encoded Packet to the underlying writer.
func (s *Subscribe) WriteTo(w io.Writer) (int64, error) {
	var buf bytes.Buffer

	//buf.Write(reserveForHeader)
	buf.Write(encodeUint16(s.MessageID))
	for _, t := range s.Subscriptions {
		buf.Write(t.Topic)
		buf.WriteByte(byte(t.Qos))
	}

	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.SUBSCRIBE, RemainingLength: buf.Len()}
	packet := fh.pack(&s.FixedHeader)
	packet.Write(buf.Bytes())
	return packet.WriteTo(w)
}

// Type returns the MQTT Packet type.
func (s *Subscribe) Type() uint8 {
	return lp.SUBSCRIBE
}

// Forwarded returns the forwarded flag.
func (s *Subscribe) Forwarded() bool {
	return s.IsForwarded
}

// String returns the name of mqtt operation.
func (s *Subscribe) String() string {
	return "sub"
}

// WriteTo writes the encoded Packet to the underlying writer.
func (s *Suback) WriteTo(w io.Writer) (int64, error) {
	var buf bytes.Buffer

	//buf.Write(reserveForHeader)
	buf.Write(encodeUint16(s.MessageID))
	for _, q := range s.Qos {
		buf.WriteByte(byte(q))
	}

	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.SUBACK, RemainingLength: buf.Len()}
	packet := fh.pack(nil)
	packet.Write(buf.Bytes())
	return packet.WriteTo(w)
}

// Type returns the MQTT Packet type.
func (s *Suback) Type() uint8 {
	return lp.SUBACK
}

// Forwarded returns the forwarded flag.
func (s *Suback) Forwarded() bool {
	return false
}

// String returns the name of mqtt operation.
func (s *Suback) String() string {
	return "suback"
}

// WriteTo writes the encoded Packet to the underlying writer.
func (u *Unsubscribe) WriteTo(w io.Writer) (int64, error) {
	var buf bytes.Buffer

	//buf.Write(reserveForHeader)
	buf.Write(encodeUint16(u.MessageID))
	for _, toptup := range u.Topics {
		buf.Write(toptup.Topic)
	}

	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.UNSUBSCRIBE, RemainingLength: buf.Len()}
	packet := fh.pack(&u.FixedHeader)
	packet.Write(buf.Bytes())
	return packet.WriteTo(w)
}

// Type returns the MQTT Packet type.
func (u *Unsubscribe) Type() uint8 {
	return lp.UNSUBSCRIBE
}

// String returns the name of mqtt operation.
func (u *Unsubscribe) String() string {
	return "unsub"
}

// WriteTo writes the encoded Packet to the underlying writer.
func (u *Unsuback) WriteTo(w io.Writer) (int64, error) {
	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.UNSUBACK, RemainingLength: 2}
	packet := fh.pack(nil)
	packet.Write(encodeUint16(u.MessageID))
	return packet.WriteTo(w)
}

// Type returns the MQTT Packet type.
func (u *Unsuback) Type() uint8 {
	return lp.UNSUBACK
}

// String returns the name of mqtt operation.
func (u *Unsuback) String() string {
	return "unsuback"
}

func unpackSubscribe(data []byte, fh FixedHeader) Packet {
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
	return &Subscribe{
		FixedHeader:   lp.FixedHeader(fh),
		MessageID:     msgID,
		Subscriptions: topics,
	}
}

func unpackSuback(data []byte, fh FixedHeader) Packet {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	var qoses []uint8
	maxlen := uint32(len(data))
	//is this efficient
	for bookmark < maxlen {
		qos := data[bookmark]
		bookmark++
		qoses = append(qoses, qos)
	}
	return &Suback{
		MessageID: msgID,
		Qos:       qoses,
	}
}

func unpackUnsubscribe(data []byte, fh FixedHeader) Packet {
	bookmark := uint32(0)
	var topics []lp.TopicQOSTuple
	msgID := readUint16(data, &bookmark)
	maxlen := uint32(len(data))
	for bookmark < maxlen {
		var t lp.TopicQOSTuple
		//		qos := data[bookmark]
		//		bookmark++
		t.Topic = readString(data, &bookmark)
		//		t.qos = uint8(qos)
		topics = append(topics, t)
	}
	return &Unsubscribe{
		FixedHeader: lp.FixedHeader(fh),
		MessageID:   msgID,
		Topics:      topics,
	}
}

func unpackUnsuback(data []byte, fh FixedHeader) Packet {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Unsuback{
		MessageID: msgID,
	}
}
