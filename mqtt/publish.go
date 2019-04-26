package mqtt

import (
	"bytes"
)

// Publish represents an MQTT publish packet.
type Publish struct {
	Header    *FixedHeader
	Topic     []byte
	MessageID uint16
	IsForwarded bool
	Payload   []byte
}

//Puback is sent for QOS level one to verify the receipt of a publish
//Qoth the spec: "A PUBACK Packet is sent by a server in response to a PUBLISH Packet from a publishing client, and by a subscriber in response to a PUBLISH Packet from the server."
type Puback struct {
	MessageID uint16
}

//Pubrec is for verifying the receipt of a publish
//Qoth the spec:"It is the second Packet of the QoS level 2 protocol flow. A PUBREC Packet is sent by the server in response to a PUBLISH Packet from a publishing client, or by a subscriber in response to a PUBLISH Packet from the server."
type Pubrec struct {
	MessageID uint16
}

//Pubrel is a response to pubrec from either the client or server.
type Pubrel struct {
	MessageID uint16
	//QOS1
	Header *FixedHeader
}

//Pubcomp is for saying is in response to a pubrel sent by the publisher
//the final member of the QOS2 flow. both sides have said "hey, we did it!"
type Pubcomp struct {
	MessageID uint16
}

// EncodeTo writes the encoded Packet to the underlying writer.
func (p *Publish) Encode() []byte {
	var buf bytes.Buffer

	buf.Write(p.Topic)
	if p.Header.QOS > 0 {
		buf.Write(encodeUint16(p.MessageID))
	}
	buf.Write(p.Payload)

	// Write to the underlying buffer
	packet := encodeParts(PUBLISH, buf.Len()+2, p.Header)
	packet.Write(varHeader[:2])
	packet.Write(buf.Bytes())
	return packet.Bytes()
}

// Type returns the MQTT Packet type.
func (p *Publish) Type() uint8 {
	return PUBLISH
}

// String returns the name of mqtt operation.
func (p *Publish) String() string {
	return "pub"
}

// EncodeTo writes the encoded Packet to the underlying writer.
func (p *Puback) Encode() []byte {
	var buf bytes.Buffer

	//buf.Write(reserveForHeader)
	buf.Write(encodeUint16(p.MessageID))

	// Write to the underlying buffer
	packet := encodeParts(PUBACK, 2, nil)
	packet.Write(buf.Bytes())
	return packet.Bytes()
}

// Type returns the MQTT Packet type.
func (p *Puback) Type() uint8 {
	return PUBACK
}

// String returns the name of mqtt operation.
func (p *Puback) String() string {
	return "puback"
}

// EncodeTo writes the encoded Packet to the underlying writer.
func (p *Pubrec) Encode() []byte {
	var buf bytes.Buffer

	//buf.Write(reserveForHeader)
	buf.Write(encodeUint16(p.MessageID))

	// Write to the underlying buffer
	packet := encodeParts(PUBREC, 2, nil)
	packet.Write(buf.Bytes())
	return packet.Bytes()
}

// Type returns the MQTT Packet type.
func (p *Pubrec) Type() uint8 {
	return PUBREC
}

// String returns the name of mqtt operation.
func (p *Pubrec) String() string {
	return "pubrec"
}

// EncodeTo writes the encoded Packet to the underlying writer.
func (p *Pubrel) Encode() []byte {
	var buf bytes.Buffer

	//buf.Write(reserveForHeader)
	buf.Write(encodeUint16(p.MessageID))

	// Write to the underlying buffer
	packet := encodeParts(PUBREL, 2, p.Header)
	packet.Write(buf.Bytes())
	return packet.Bytes()
}

// Type returns the MQTT Packet type.
func (p *Pubrel) Type() uint8 {
	return PUBREL
}

// String returns the name of mqtt operation.
func (p *Pubrel) String() string {
	return "pubrel"
}

// EncodeTo writes the encoded Packet to the underlying writer.
func (p *Pubcomp) Encode() []byte {
	var buf bytes.Buffer

	//buf.Write(reserveForHeader)
	buf.Write(encodeUint16(p.MessageID))

	// Write to the underlying buffer
	packet := encodeParts(PUBCOMP, 2, nil)
	packet.Write(buf.Bytes())
	return packet.Bytes()
}

// Type returns the MQTT Packet type.
func (p *Pubcomp) Type() uint8 {
	return PUBCOMP
}

// String returns the name of mqtt operation.
func (p *Pubcomp) String() string {
	return "pubcomp"
}

func unpackPublish(data []byte, hdr *FixedHeader) Packet {
	bookmark := uint32(0)
	topic := readString(data, &bookmark)
	var msgID uint16
	if hdr.QOS > 0 {
		msgID = readUint16(data, &bookmark)
	}

	return &Publish{
		Topic:     topic,
		Header:    hdr,
		Payload:   data[bookmark:],
		MessageID: msgID,
	}
}

func unpackPuback(data []byte, hdr *FixedHeader) Packet {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Puback{
		MessageID: msgID,
	}
}

func unpackPubrec(data []byte, hdr *FixedHeader) Packet {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Pubrec{
		MessageID: msgID,
	}
}

func unpackPubrel(data []byte, hdr *FixedHeader) Packet {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Pubrel{
		Header:    hdr,
		MessageID: msgID,
	}
}

func unpackPubcomp(data []byte, hdr *FixedHeader) Packet {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Pubcomp{
		MessageID: msgID,
	}
}
