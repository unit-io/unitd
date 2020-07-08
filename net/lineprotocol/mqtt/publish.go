package mqtt

import (
	"bytes"
	"errors"
	"io"

	lp "github.com/unit-io/unitd/net/lineprotocol"
)

type (
	Publish lp.Publish
	Puback  lp.Puback
	Pubrec  lp.Pubrec
	Pubrel  lp.Pubrel
	Pubcomp lp.Pubcomp
)

func (p *Publish) encode() (bytes.Buffer, error) {
	var buf bytes.Buffer
	var length int
	length = 2 + len(p.Topic) + len(p.Payload)
	if p.FixedHeader.Qos > 0 {
		length += 2
	}
	if length > lp.MaxMessageSize {
		return buf, errors.New("Message too large")
	}

	buf.Write(encodeBytes(p.Topic))
	if p.FixedHeader.Qos > 0 {
		buf.Write(encodeUint16(p.MessageID))
	}
	buf.Write(p.Payload)
	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.PUBLISH, RemainingLength: length}
	packet := fh.pack(&p.FixedHeader)
	packet.Write(buf.Bytes())
	return packet, nil
}

// Encode encodes message into binary data
func (p *Publish) Encode() []byte {
	buf, err := p.encode()
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

// WriteTo writes the encoded message to the buffer.
func (p *Publish) WriteTo(w io.Writer) (int64, error) {
	buf, err := p.encode()
	if err != nil {
		return 0, err
	}
	return buf.WriteTo(w)
}

// Type returns the MQTT Packet type.
func (p *Publish) Type() uint8 {
	return lp.PUBLISH
}

// String returns the name of mqtt operation.
func (p *Publish) String() string {
	return "pub"
}

// Info returns Qos and MessageID of this packet.
func (p *Publish) Info() lp.Info {
	return lp.Info{Qos: p.Qos, MessageID: p.MessageID}
}

func (p *Puback) encode() (bytes.Buffer, error) {
	fh := FixedHeader{MessageType: lp.PUBACK, RemainingLength: 2}
	packet := fh.pack(nil)
	_, err := packet.Write(encodeUint16(p.MessageID))
	return packet, err
}

// Encode encodes message into binary data
func (p *Puback) Encode() []byte {
	buf, err := p.encode()
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

// WriteTo writes the encoded Packet to the underlying writer.
func (p *Puback) WriteTo(w io.Writer) (int64, error) {
	buf, err := p.encode()
	if err != nil {
		return 0, err
	}
	return buf.WriteTo(w)
}

// Type returns the MQTT Packet type.
func (p *Puback) Type() uint8 {
	return lp.PUBACK
}

// String returns the name of mqtt operation.
func (p *Puback) String() string {
	return "puback"
}

// Info returns Qos and MessageID of this packet.
func (p *Puback) Info() lp.Info {
	return lp.Info{Qos: 0, MessageID: p.MessageID}
}

func (p *Pubrec) encode() (bytes.Buffer, error) {
	fh := FixedHeader{MessageType: lp.PUBREC, RemainingLength: 2}
	packet := fh.pack(nil)
	_, err := packet.Write(encodeUint16(p.MessageID))
	return packet, err
}

// Encode encodes message into binary data
func (p *Pubrec) Encode() []byte {
	buf, err := p.encode()
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

// WriteTo writes the encoded Packet to the underlying writer.
func (p *Pubrec) WriteTo(w io.Writer) (int64, error) {
	buf, err := p.encode()
	if err != nil {
		return 0, err
	}
	return buf.WriteTo(w)
}

// Type returns the MQTT Packet type.
func (p *Pubrec) Type() uint8 {
	return lp.PUBREC
}

// String returns the name of mqtt operation.
func (p *Pubrec) String() string {
	return "pubrec"
}

// Info returns Qos and MessageID of this packet.
func (p *Pubrec) Info() lp.Info {
	return lp.Info{Qos: 0, MessageID: p.MessageID}
}

func (p *Pubrel) encode() (bytes.Buffer, error) {
	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.PUBREL, RemainingLength: 2}
	packet := fh.pack(&p.FixedHeader)
	_, err := packet.Write(encodeUint16(p.MessageID))
	return packet, err
}

// Encode encodes message into binary data
func (p *Pubrel) Encode() []byte {
	buf, err := p.encode()
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

// WriteTo writes the encoded Packet to the underlying writer.
func (p *Pubrel) WriteTo(w io.Writer) (int64, error) {
	buf, err := p.encode()
	if err != nil {
		return 0, err
	}
	return buf.WriteTo(w)
}

// Type returns the MQTT Packet type.
func (p *Pubrel) Type() uint8 {
	return lp.PUBREL
}

// String returns the name of mqtt operation.
func (p *Pubrel) String() string {
	return "pubrel"
}

// Info returns Qos and MessageID of this packet.
func (p *Pubrel) Info() lp.Info {
	return lp.Info{Qos: p.Qos, MessageID: p.MessageID}
}

func (p *Pubcomp) encode() (bytes.Buffer, error) {
	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.PUBCOMP, RemainingLength: 2}
	packet := fh.pack(nil)
	_, err := packet.Write(encodeUint16(p.MessageID))
	return packet, err
}

// Encode encodes message into binary data
func (p *Pubcomp) Encode() []byte {
	buf, err := p.encode()
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

// WriteTo writes the encoded Packet to the underlying writer.
func (p *Pubcomp) WriteTo(w io.Writer) (int64, error) {
	buf, err := p.encode()
	if err != nil {
		return 0, err
	}
	return buf.WriteTo(w)
}

// Type returns the MQTT Packet type.
func (p *Pubcomp) Type() uint8 {
	return lp.PUBCOMP
}

// String returns the name of mqtt operation.
func (p *Pubcomp) String() string {
	return "pubcomp"
}

// Info returns Qos and MessageID of this packet.
func (p *Pubcomp) Info() lp.Info {
	return lp.Info{Qos: 0, MessageID: p.MessageID}
}

func unpackPublish(data []byte, fh FixedHeader) Packet {
	bookmark := uint32(0)
	topic := readString(data, &bookmark)
	var msgID uint16
	if fh.Qos > 0 {
		msgID = readUint16(data, &bookmark)
	}

	return &Publish{
		FixedHeader: lp.FixedHeader(fh),
		Topic:       topic,
		Payload:     data[bookmark:],
		MessageID:   msgID,
	}
}

func unpackPuback(data []byte, fh FixedHeader) Packet {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Puback{
		MessageID: msgID,
	}
}

func unpackPubrec(data []byte, fh FixedHeader) Packet {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Pubrec{
		MessageID: msgID,
	}
}

func unpackPubrel(data []byte, fh FixedHeader) Packet {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Pubrel{
		FixedHeader: lp.FixedHeader(fh),
		MessageID:   msgID,
	}
}

func unpackPubcomp(data []byte, fh FixedHeader) Packet {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Pubcomp{
		MessageID: msgID,
	}
}
