package grpc

import (
	"bytes"

	"github.com/golang/protobuf/proto"
	lp "github.com/unit-io/unitd/lineprotocol"
	pbx "github.com/unit-io/unitd/proto"
)

func encodeSubscribe(s lp.Subscribe) (bytes.Buffer, error) {
	var msg bytes.Buffer
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
		return msg, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_SUBSCRIBE, RemainingLength: uint32(len(pkt))}
	msg = fh.pack()
	_, err = msg.Write(pkt)
	return msg, err
}

func encodeSuback(s lp.Suback) (bytes.Buffer, error) {
	var msg bytes.Buffer
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
		return msg, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_SUBACK, RemainingLength: uint32(len(pkt))}
	msg = fh.pack()
	_, err = msg.Write(pkt)
	return msg, err
}

func encodeUnsubscribe(u lp.Unsubscribe) (bytes.Buffer, error) {
	var msg bytes.Buffer
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
		return msg, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_UNSUBSCRIBE, RemainingLength: uint32(len(pkt))}
	msg = fh.pack()
	_, err = msg.Write(pkt)
	return msg, err
}

func encodeUnsuback(u lp.Unsuback) (bytes.Buffer, error) {
	var msg bytes.Buffer
	unusuback := pbx.Unsuback{
		MessageID: uint32(u.MessageID),
	}
	pkt, err := proto.Marshal(&unusuback)
	if err != nil {
		return msg, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_UNSUBACK, RemainingLength: uint32(len(pkt))}
	msg = fh.pack()
	_, err = msg.Write(pkt)
	return msg, err
}

func unpackSubscribe(data []byte) lp.Packet {
	var pkt pbx.Subscribe
	proto.Unmarshal(data, &pkt)
	var topics []lp.TopicQOSTuple
	for _, sub := range pkt.Subscribers {
		var t lp.TopicQOSTuple
		t.Topic = sub.Topic
		t.Qos = uint8(sub.Qos)
		topics = append(topics, t)
	}
	return &lp.Subscribe{
		MessageID:     uint16(pkt.MessageID),
		Subscriptions: topics,
	}
}

func unpackSuback(data []byte) lp.Packet {
	var pkt pbx.Suback
	proto.Unmarshal(data, &pkt)
	var qoses []uint8
	//is this efficient
	for _, qos := range pkt.Qos {
		qoses = append(qoses, uint8(qos))
	}
	return &lp.Suback{
		MessageID: uint16(pkt.MessageID),
		Qos:       qoses,
	}
}

func unpackUnsubscribe(data []byte) lp.Packet {
	var pkt pbx.Unsubscribe
	proto.Unmarshal(data, &pkt)
	var topics []lp.TopicQOSTuple
	for _, sub := range pkt.Subscribers {
		var t lp.TopicQOSTuple
		t.Topic = sub.Topic
		t.Qos = uint8(sub.Qos)
		topics = append(topics, t)
	}
	return &lp.Unsubscribe{
		MessageID: uint16(pkt.MessageID),
		Topics:    topics,
	}
}

func unpackUnsuback(data []byte) lp.Packet {
	var pkt pbx.Unsuback
	proto.Unmarshal(data, &pkt)
	return &lp.Unsuback{
		MessageID: uint16(pkt.MessageID),
	}
}
