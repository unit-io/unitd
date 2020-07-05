package grpc

import (
	"bytes"
	"io"

	"github.com/golang/protobuf/proto"
	lp "github.com/unit-io/unitd/net/lineprotocol"
	pbx "github.com/unit-io/unitd/proto"
)

type (
	Subscribe   lp.Subscribe
	Suback      lp.Suback
	Unsubscribe lp.Unsubscribe
	Unsuback    lp.Unsuback
)

func (s *Subscribe) encode() (bytes.Buffer, error) {
	var buf bytes.Buffer
	var subs []*pbx.Subscriber
	for _, t := range s.Subscriptions {
		var sub *pbx.Subscriber
		sub.Topic = t.Topic
		sub.Qos = uint32(t.Qos)
		subs = append(subs, sub)
	}
	sub := pbx.Subscribe{
		MessageID:   uint32(s.MessageID),
		Subscribers: subs,
	}
	pkt, err := proto.Marshal(&sub)
	if err != nil {
		return buf, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_SUBSCRIBE, RemainingLength: uint32(len(pkt))}
	buf = fh.pack()
	_, err = buf.Write(pkt)
	return buf, err
}

// Encode encodes message into binary data
func (s *Subscribe) Encode() []byte {
	buf, err := s.encode()
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

// WriteTo writes the encoded Packet to the underlying writer.
func (s *Subscribe) WriteTo(w io.Writer) (int64, error) {
	buf, err := s.encode()
	if err != nil {
		return 0, err
	}
	return buf.WriteTo(w)
}

// Type returns the Packet type.
func (s *Subscribe) Type() uint8 {
	return uint8(pbx.MessageType_SUBSCRIBE)
}

// String returns the name of operation.
func (s *Subscribe) String() string {
	return "sub"
}

// Info returns Qos and MessageID of this packet.
func (s *Subscribe) Info() lp.Info {
	return lp.Info{Qos: 1, MessageID: s.MessageID}
}

func (s *Suback) encode() (bytes.Buffer, error) {
	var buf bytes.Buffer
	var qoss []uint32
	for _, q := range s.Qos {
		qoss = append(qoss, uint32(q))
	}
	suback := pbx.Suback{
		MessageID: uint32(s.MessageID),
		Qos:       qoss,
	}
	pkt, err := proto.Marshal(&suback)
	if err != nil {
		return buf, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_SUBACK, RemainingLength: uint32(len(pkt))}
	buf = fh.pack()
	_, err = buf.Write(pkt)
	return buf, err
}

// Encode encodes message into binary data
func (s *Suback) Encode() []byte {
	buf, err := s.encode()
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

// WriteTo writes the encoded Packet to the underlying writer.
func (s *Suback) WriteTo(w io.Writer) (int64, error) {
	buf, err := s.encode()
	if err != nil {
		return 0, err
	}
	return buf.WriteTo(w)
}

// Type returns the Packet type.
func (s *Suback) Type() uint8 {
	return uint8(pbx.MessageType_SUBACK)
}

// String returns the name of operation.
func (s *Suback) String() string {
	return "suback"
}

// Info returns Qos and MessageID of this packet.
func (s *Suback) Info() lp.Info {
	return lp.Info{Qos: 0, MessageID: s.MessageID}
}

func (u *Unsubscribe) encode() (bytes.Buffer, error) {
	var buf bytes.Buffer
	var subs []*pbx.Subscriber
	for _, t := range u.Topics {
		var sub *pbx.Subscriber
		sub.Topic = t.Topic
		sub.Qos = uint32(t.Qos)
		subs = append(subs, sub)
	}
	unsub := pbx.Unsubscribe{
		MessageID:   uint32(u.MessageID),
		Subscribers: subs,
	}
	pkt, err := proto.Marshal(&unsub)
	if err != nil {
		return buf, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_UNSUBSCRIBE, RemainingLength: uint32(len(pkt))}
	buf = fh.pack()
	_, err = buf.Write(pkt)
	return buf, err
}

// Encode encodes message into binary data
func (u *Unsubscribe) Encode() []byte {
	buf, err := u.encode()
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

// Write writes the encoded Packet to the underlying writer.
func (u *Unsubscribe) WriteTo(w io.Writer) (int64, error) {
	buf, err := u.encode()
	if err != nil {
		return 0, err
	}
	return buf.WriteTo(w)
}

// Type returns the Packet type.
func (u *Unsubscribe) Type() uint8 {
	return uint8(pbx.MessageType_UNSUBSCRIBE)
}

// String returns the name of operation.
func (u *Unsubscribe) String() string {
	return "unsub"
}

// Info returns Qos and MessageID of this packet.
func (u *Unsubscribe) Info() lp.Info {
	return lp.Info{Qos: 1, MessageID: u.MessageID}
}

func (u *Unsuback) encode() (bytes.Buffer, error) {
	var buf bytes.Buffer
	unusuback := pbx.Unsuback{
		MessageID: uint32(u.MessageID),
	}
	pkt, err := proto.Marshal(&unusuback)
	if err != nil {
		return buf, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_UNSUBACK, RemainingLength: uint32(len(pkt))}
	buf = fh.pack()
	_, err = buf.Write(pkt)
	return buf, err
}

// Encode encodes message into binary data
func (u *Unsuback) Encode() []byte {
	buf, err := u.encode()
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

// WriteTo writes the encoded Packet to the underlying writer.
func (u *Unsuback) WriteTo(w io.Writer) (int64, error) {
	buf, err := u.encode()
	if err != nil {
		return 0, err
	}
	return buf.WriteTo(w)
}

// Type returns the Packet type.
func (u *Unsuback) Type() uint8 {
	return uint8(pbx.MessageType_UNSUBACK)
}

// String returns the name of operation.
func (u *Unsuback) String() string {
	return "unsuback"
}

// Info returns Qos and MessageID of this packet.
func (u *Unsuback) Info() lp.Info {
	return lp.Info{Qos: 0, MessageID: u.MessageID}
}

func unpackSubscribe(data []byte) Packet {
	var pkt pbx.Subscribe
	proto.Unmarshal(data, &pkt)
	var topics []lp.TopicQOSTuple
	for _, sub := range pkt.Subscribers {
		var t lp.TopicQOSTuple
		t.Topic = sub.Topic
		t.Qos = uint8(sub.Qos)
		topics = append(topics, t)
	}
	return &Subscribe{
		MessageID:     uint16(pkt.MessageID),
		Subscriptions: topics,
	}
}

func unpackSuback(data []byte) Packet {
	var pkt pbx.Suback
	proto.Unmarshal(data, &pkt)
	var qoses []uint8
	//is this efficient
	for _, qos := range pkt.Qos {
		qoses = append(qoses, uint8(qos))
	}
	return &Suback{
		MessageID: uint16(pkt.MessageID),
		Qos:       qoses,
	}
}

func unpackUnsubscribe(data []byte) Packet {
	var pkt pbx.Unsubscribe
	proto.Unmarshal(data, &pkt)
	var topics []lp.TopicQOSTuple
	for _, sub := range pkt.Subscribers {
		var t lp.TopicQOSTuple
		t.Topic = sub.Topic
		t.Qos = uint8(sub.Qos)
		topics = append(topics, t)
	}
	return &Unsubscribe{
		MessageID: uint16(pkt.MessageID),
		Topics:    topics,
	}
}

func unpackUnsuback(data []byte) Packet {
	var pkt pbx.Unsuback
	proto.Unmarshal(data, &pkt)
	return &Unsuback{
		MessageID: uint16(pkt.MessageID),
	}
}
