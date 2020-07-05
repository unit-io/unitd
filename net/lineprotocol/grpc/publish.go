package grpc

import (
	"bytes"
	"io"

	"github.com/golang/protobuf/proto"
	lp "github.com/unit-io/unitd/net/lineprotocol"
	pbx "github.com/unit-io/unitd/proto"
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
	pub := pbx.Publish{
		MessageID: uint32(p.MessageID),
		Topic:     p.Topic,
		Payload:   p.Payload,
	}
	pkt, err := proto.Marshal(&pub)
	if err != nil {
		return buf, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_PUBLISH, RemainingLength: uint32(len(pkt))}
	buf = fh.pack()
	_, err = buf.Write(pkt)
	return buf, err
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

// Type returns the packet type.
func (p *Publish) Type() uint8 {
	return uint8(pbx.MessageType_PUBLISH)
}

// String returns the name of operation.
func (p *Publish) String() string {
	return "pub"
}

// Info returns Qos and MessageID of this packet.
func (p *Publish) Info() lp.Info {
	return lp.Info{Qos: p.Qos, MessageID: p.MessageID}
}

func (p *Puback) encode() (bytes.Buffer, error) {
	var buf bytes.Buffer
	puback := pbx.Puback{
		MessageID: uint32(p.MessageID),
	}
	pkt, err := proto.Marshal(&puback)
	if err != nil {
		return buf, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_PUBACK, RemainingLength: uint32(len(pkt))}
	buf = fh.pack()
	_, err = buf.Write(pkt)
	return buf, err
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

// Type returns the Packet type.
func (p *Puback) Type() uint8 {
	return uint8(pbx.MessageType_PUBACK)
}

// String returns the name of operation.
func (p *Puback) String() string {
	return "puback"
}

// Info returns Qos and MessageID of this packet.
func (p *Puback) Info() lp.Info {
	return lp.Info{Qos: 0, MessageID: p.MessageID}
}

func (p *Pubrec) encode() (bytes.Buffer, error) {
	var buf bytes.Buffer
	pubrec := pbx.Pubrec{
		MessageID: uint32(p.MessageID),
	}
	pkt, err := proto.Marshal(&pubrec)
	if err != nil {
		return buf, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_PUBREC, RemainingLength: uint32(len(pkt))}
	buf = fh.pack()
	_, err = buf.Write(pkt)
	return buf, err
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

// Type returns the Packet type.
func (p *Pubrec) Type() uint8 {
	return uint8(pbx.MessageType_PUBREC)
}

// String returns the name of operation.
func (p *Pubrec) String() string {
	return "pubrec"
}

// Info returns Qos and MessageID of this packet.
func (p *Pubrec) Info() lp.Info {
	return lp.Info{Qos: 0, MessageID: p.MessageID}
}

func (p *Pubrel) encode() (bytes.Buffer, error) {
	var buf bytes.Buffer
	pubrel := pbx.Pubrel{
		MessageID: uint32(p.MessageID),
	}
	pkt, err := proto.Marshal(&pubrel)
	if err != nil {
		return buf, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_PUBREL, RemainingLength: uint32(len(pkt))}
	buf = fh.pack()
	_, err = buf.Write(pkt)
	return buf, err
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

// Type returns the Packet type.
func (p *Pubrel) Type() uint8 {
	return uint8(pbx.MessageType_PUBREL)
}

// String returns the name of operation.
func (p *Pubrel) String() string {
	return "pubrel"
}

// Info returns Qos and MessageID of this packet.
func (p *Pubrel) Info() lp.Info {
	return lp.Info{Qos: p.Qos, MessageID: p.MessageID}
}

func (p *Pubcomp) encode() (bytes.Buffer, error) {
	var buf bytes.Buffer
	pubcomp := pbx.Pubcomp{
		MessageID: uint32(p.MessageID),
	}
	pkt, err := proto.Marshal(&pubcomp)
	if err != nil {
		return buf, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_PUBCOMP, RemainingLength: uint32(len(pkt))}
	buf = fh.pack()
	_, err = buf.Write(pkt)
	return buf, err
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

// Type returns the Packet type.
func (p *Pubcomp) Type() uint8 {
	return uint8(pbx.MessageType_PUBCOMP)
}

// String returns the name of operation.
func (p *Pubcomp) String() string {
	return "pubcomp"
}

// Info returns Qos and MessageID of this packet.
func (p *Pubcomp) Info() lp.Info {
	return lp.Info{Qos: 0, MessageID: p.MessageID}
}

func unpackPublish(data []byte) Packet {
	var pkt pbx.Publish
	proto.Unmarshal(data, &pkt)

	fh := lp.FixedHeader{
		Qos: uint8(pkt.Qos),
	}

	return &Publish{
		FixedHeader: fh,
		MessageID:   uint16(pkt.MessageID),
		Topic:       pkt.Topic,
		Payload:     pkt.Payload,
	}
}

func unpackPuback(data []byte) Packet {
	var pkt pbx.Puback
	proto.Unmarshal(data, &pkt)
	return &Puback{
		MessageID: uint16(pkt.MessageID),
	}
}

func unpackPubrec(data []byte) Packet {
	var pkt pbx.Pubrec
	proto.Unmarshal(data, &pkt)
	return &Pubrec{
		MessageID: uint16(pkt.MessageID),
	}
}

func unpackPubrel(data []byte) Packet {
	var pkt pbx.Pubrel
	proto.Unmarshal(data, &pkt)

	fh := lp.FixedHeader{
		Qos: uint8(pkt.Qos),
	}

	return &Pubrel{
		FixedHeader: fh,
		MessageID:   uint16(pkt.MessageID),
	}
}

func unpackPubcomp(data []byte) Packet {
	var pkt pbx.Pubcomp
	proto.Unmarshal(data, &pkt)
	return &Pubcomp{
		MessageID: uint16(pkt.MessageID),
	}
}
