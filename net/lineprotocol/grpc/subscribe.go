package grpc

import (
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

// WriteTo writes the encoded Packet to the underlying writer.
func (s *Subscribe) WriteTo(w io.Writer) (int64, error) {
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
		return 0, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_SUBSCRIBE, RemainingLength: uint32(len(pkt))}
	buf := fh.pack()
	buf.Write(pkt)
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

// WriteTo writes the encoded Packet to the underlying writer.
func (s *Suback) WriteTo(w io.Writer) (int64, error) {
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
		return 0, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_SUBACK, RemainingLength: uint32(len(pkt))}
	buf := fh.pack()
	buf.Write(pkt)
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

// WriteTo writes the encoded Packet to the underlying writer.
func (u *Unsubscribe) WriteTo(w io.Writer) (int64, error) {
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
		return 0, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_UNSUBSCRIBE, RemainingLength: uint32(len(pkt))}
	buf := fh.pack()
	buf.Write(pkt)
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

// WriteTo writes the encoded Packet to the underlying writer.
func (u *Unsuback) WriteTo(w io.Writer) (int64, error) {
	unusuback := pbx.Unsuback{
		MessageID: uint32(u.MessageID),
	}
	pkt, err := proto.Marshal(&unusuback)
	if err != nil {
		return 0, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_UNSUBACK, RemainingLength: uint32(len(pkt))}
	buf := fh.pack()
	buf.Write(pkt)
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
