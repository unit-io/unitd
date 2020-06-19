package mqtt

import (
	"bytes"
	"io"

	lp "github.com/unit-io/unitd/net"
)

type (
	Publish lp.Publish
	Puback  lp.Puback
	Pubrec  lp.Pubrec
	Pubrel  lp.Pubrel
	Pubcomp lp.Pubcomp
)

// WriteTo writes the encoded Packet to the underlying writer.
func (p *Publish) WriteTo(w io.Writer) (int64, error) {
	var buf bytes.Buffer

	buf.Write(p.Topic)
	if p.FixedHeader.QOS > 0 {
		buf.Write(encodeUint16(p.MessageID))
	}
	buf.Write(p.Payload)

	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.PUBLISH, RemainingLength: buf.Len() + 2}
	packet := fh.pack(&p.FixedHeader)
	packet.Write(varHeader[:2])
	packet.Write(buf.Bytes())
	return packet.WriteTo(w)
}

// Type returns the MQTT Packet type.
func (p *Publish) Type() uint8 {
	return lp.PUBLISH
}

// String returns the name of mqtt operation.
func (p *Publish) String() string {
	return "pub"
}

// WriteTo writes the encoded Packet to the underlying writer.
func (p *Puback) WriteTo(w io.Writer) (int64, error) {
	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.PUBACK, RemainingLength: 2}
	packet := fh.pack(nil)
	packet.Write(encodeUint16(p.MessageID))
	return packet.WriteTo(w)
}

// Type returns the MQTT Packet type.
func (p *Puback) Type() uint8 {
	return lp.PUBACK
}

// String returns the name of mqtt operation.
func (p *Puback) String() string {
	return "puback"
}

// WriteTo writes the encoded Packet to the underlying writer.
func (p *Pubrec) WriteTo(w io.Writer) (int64, error) {
	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.PUBREC, RemainingLength: 2}
	packet := fh.pack(nil)
	packet.Write(encodeUint16(p.MessageID))
	return packet.WriteTo(w)
}

// Type returns the MQTT Packet type.
func (p *Pubrec) Type() uint8 {
	return lp.PUBREC
}

// String returns the name of mqtt operation.
func (p *Pubrec) String() string {
	return "pubrec"
}

// WriteTo writes the encoded Packet to the underlying writer.
func (p *Pubrel) WriteTo(w io.Writer) (int64, error) {
	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.PUBREL, RemainingLength: 2}
	packet := fh.pack(&p.FixedHeader)
	packet.Write(encodeUint16(p.MessageID))
	return packet.WriteTo(w)
}

// Type returns the MQTT Packet type.
func (p *Pubrel) Type() uint8 {
	return lp.PUBREL
}

// String returns the name of mqtt operation.
func (p *Pubrel) String() string {
	return "pubrel"
}

// WriteTo writes the encoded Packet to the underlying writer.
func (p *Pubcomp) WriteTo(w io.Writer) (int64, error) {
	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.PUBCOMP, RemainingLength: 2}
	packet := fh.pack(nil)
	packet.Write(encodeUint16(p.MessageID))
	return packet.WriteTo(w)
}

// Type returns the MQTT Packet type.
func (p *Pubcomp) Type() uint8 {
	return lp.PUBCOMP
}

// String returns the name of mqtt operation.
func (p *Pubcomp) String() string {
	return "pubcomp"
}

func unpackPublish(data []byte, fh FixedHeader) Packet {
	bookmark := uint32(0)
	topic := readString(data, &bookmark)
	var msgID uint16
	if fh.QOS > 0 {
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
