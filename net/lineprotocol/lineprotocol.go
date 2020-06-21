package lineprotocol

import (
	"fmt"
	"io"
)

//Packet is the interface all our packets in the line protocol will be implementing
type Packet interface {
	fmt.Stringer

	Type() uint8
	WriteTo(io.Writer) (int64, error)
}

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
	MessageType     byte
	DUP             bool
	Retain          bool
	QOS             uint8
	RemainingLength int
}

// Connect represents a connect packet.
type Connect struct {
	ProtoName      []byte
	Version        uint8
	InsecureFlag   bool
	UsernameFlag   bool
	PasswordFlag   bool
	WillRetainFlag bool
	WillQOS        uint8
	WillFlag       bool
	CleanSessFlag  bool
	KeepAlive      uint16
	ClientID       []byte
	WillTopic      []byte
	WillMessage    []byte
	Username       []byte
	Password       []byte

	Packet
}

// Connack represents an connack packet.
// 0x00 connection accepted
// 0x01 refused: unacceptable proto version
// 0x02 refused: identifier rejected
// 0x03 refused server unavailiable
// 0x04 bad user or password
// 0x05 not authorized
type Connack struct {
	ReturnCode uint8
	Packet
}

//Pingreq is a keepalive
type Pingreq struct {
}

//Pingresp is for saying "hey, the server is alive"
type Pingresp struct {
}

//Disconnect is to signal you want to cease communications with the server
type Disconnect struct {
}

// Publish represents a publish packet.
type Publish struct {
	FixedHeader
	Topic       []byte
	MessageID   uint16
	IsForwarded bool
	Payload     []byte

	Packet
}

//Puback is sent for QOS level one to verify the receipt of a publish
//Qoth the spec: "A PUBACK Packet is sent by a server in response to a PUBLISH Packet from a publishing client, and by a subscriber in response to a PUBLISH Packet from the server."
type Puback struct {
	MessageID uint16

	Packet
}

//Pubrec is for verifying the receipt of a publish
//Qoth the spec:"It is the second Packet of the QoS level 2 protocol flow. A PUBREC Packet is sent by the server in response to a PUBLISH Packet from a publishing client, or by a subscriber in response to a PUBLISH Packet from the server."
type Pubrec struct {
	MessageID uint16

	Packet
}

//Pubrel is a response to pubrec from either the client or server.
type Pubrel struct {
	FixedHeader
	MessageID uint16

	Packet
}

//Pubcomp is for saying is in response to a pubrel sent by the publisher
//the final member of the QOS2 flow. both sides have said "hey, we did it!"
type Pubcomp struct {
	MessageID uint16

	Packet
}

//TopicQOSTuple is a struct for pairing the Qos and topic together
//for the QOS' pairs in unsubscribe and subscribe
type TopicQOSTuple struct {
	Qos   uint8
	Topic []byte
}

//Subscribe tells the server which topics the client would like to subscribe to
type Subscribe struct {
	FixedHeader
	MessageID     uint16
	IsForwarded   bool
	Subscriptions []TopicQOSTuple

	Packet
}

//Suback is to say "hey, you got it buddy. I will send you messages that fit this pattern"
type Suback struct {
	MessageID uint16
	Qos       []uint8

	Packet
}

//Unsubscribe is the Packet to send if you don't want to subscribe to a topic anymore
type Unsubscribe struct {
	FixedHeader
	MessageID   uint16
	IsForwarded bool
	Topics      []TopicQOSTuple

	Packet
}

//Unsuback is to unsubscribe as suback is to subscribe
type Unsuback struct {
	MessageID uint16

	Packet
}
