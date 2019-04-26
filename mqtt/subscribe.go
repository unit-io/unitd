package mqtt

import (
	"bytes"
)

//TopicQOSTuple is a struct for pairing the Qos and topic together
//for the QOS' pairs in unsubscribe and subscribe
type TopicQOSTuple struct {
	Qos   uint8
	Topic []byte
}

//Subscribe tells the server which topics the client would like to subscribe to
type Subscribe struct {
	Header        *FixedHeader
	MessageID     uint16
	IsForwarded bool
	Subscriptions []TopicQOSTuple
}

//Suback is to say "hey, you got it buddy. I will send you messages that fit this pattern"
type Suback struct {
	MessageID uint16
	Qos       []uint8
}

//Unsubscribe is the Packet to send if you don't want to subscribe to a topic anymore
type Unsubscribe struct {
	Header    *FixedHeader
	MessageID uint16
	IsForwarded bool
	Topics    []TopicQOSTuple
}

//Unsuback is to unsubscribe as suback is to subscribe
type Unsuback struct {
	MessageID uint16
}

// Write writes the encoded Packet to the underlying writer.
func (s *Subscribe) Encode() []byte {
	var buf bytes.Buffer

	//buf.Write(reserveForHeader)
	buf.Write(encodeUint16(s.MessageID))
	for _, t := range s.Subscriptions {
		buf.Write(t.Topic)
		buf.WriteByte(byte(t.Qos))
	}

	// Write to the underlying buffer
	packet := encodeParts(SUBSCRIBE, buf.Len(), s.Header)
	packet.Write(buf.Bytes())
	return packet.Bytes()
}

// Type returns the MQTT Packet type.
func (s *Subscribe) Type() uint8 {
	return SUBSCRIBE
}

// String returns the name of mqtt operation.
func (s *Subscribe) String() string {
	return "sub"
}

// Write writes the encoded Packet to the underlying writer.
func (s *Suback) Encode() []byte {
	var buf bytes.Buffer

	//buf.Write(reserveForHeader)
	buf.Write(encodeUint16(s.MessageID))
	for _, q := range s.Qos {
		buf.WriteByte(byte(q))
	}

	// Write to the underlying buffer
	packet := encodeParts(SUBACK, buf.Len(), nil)
	packet.Write(buf.Bytes())
	return packet.Bytes()
}

// Type returns the MQTT Packet type.
func (s *Suback) Type() uint8 {
	return SUBACK
}

// String returns the name of mqtt operation.
func (s *Suback) String() string {
	return "suback"
}

// Write writes the encoded Packet to the underlying writer.
func (u *Unsubscribe) Encode() []byte {
	var buf bytes.Buffer

	//buf.Write(reserveForHeader)
	buf.Write(encodeUint16(u.MessageID))
	for _, toptup := range u.Topics {
		buf.Write(toptup.Topic)
	}

	// Write to the underlying buffer
	packet := encodeParts(UNSUBSCRIBE, buf.Len(), u.Header)
	packet.Write(buf.Bytes())
	return packet.Bytes()
}

// Type returns the MQTT Packet type.
func (u *Unsubscribe) Type() uint8 {
	return UNSUBSCRIBE
}

// String returns the name of mqtt operation.
func (u *Unsubscribe) String() string {
	return "unsub"
}

// Write writes the encoded Packet to the underlying writer.
func (u *Unsuback) Encode() []byte {
	var buf bytes.Buffer

	//buf.Write(reserveForHeader)
	buf.Write(encodeUint16(u.MessageID))

	// Write to the underlying buffer
	packet := encodeParts(UNSUBACK, 2, nil)
	packet.Write(buf.Bytes())
	return packet.Bytes()
}

// Type returns the MQTT Packet type.
func (u *Unsuback) Type() uint8 {
	return UNSUBACK
}

// String returns the name of mqtt operation.
func (u *Unsuback) String() string {
	return "unsuback"
}

func unpackSubscribe(data []byte, hdr *FixedHeader) Packet {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	var topics []TopicQOSTuple
	maxlen := uint32(len(data))
	for bookmark < maxlen {
		var t TopicQOSTuple
		t.Topic = readString(data, &bookmark)
		qos := data[bookmark]
		bookmark++
		t.Qos = uint8(qos)
		topics = append(topics, t)
	}
	return &Subscribe{
		Header:        hdr,
		MessageID:     msgID,
		Subscriptions: topics,
	}
}

func unpackSuback(data []byte, hdr *FixedHeader) Packet {
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

func unpackUnsubscribe(data []byte, hdr *FixedHeader) Packet {
	bookmark := uint32(0)
	var topics []TopicQOSTuple
	msgID := readUint16(data, &bookmark)
	maxlen := uint32(len(data))
	for bookmark < maxlen {
		var t TopicQOSTuple
		//		qos := data[bookmark]
		//		bookmark++
		t.Topic = readString(data, &bookmark)
		//		t.qos = uint8(qos)
		topics = append(topics, t)
	}
	return &Unsubscribe{
		Header:    hdr,
		MessageID: msgID,
		Topics:    topics,
	}
}

func unpackUnsuback(data []byte, hdr *FixedHeader) Packet {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Unsuback{
		MessageID: msgID,
	}
}
