package grpc

import (
	"io"

	"github.com/golang/protobuf/proto"
	lp "github.com/unit-io/unitd/net"
	pbx "github.com/unit-io/unitd/proto"
)

type (
	Publish lp.Publish
	Puback  lp.Puback
	Pubrec  lp.Pubrec
	Pubrel  lp.Pubrel
	Pubcomp lp.Pubcomp
)

// Encode writes the encoded message to the buffer.
func (p *Publish) WriteTo(w io.Writer) (int64, error) {
	pub := pbx.Publish{
		MessageID: uint32(p.MessageID),
		Topic:     p.Topic,
		Payload:   p.Payload,
	}
	pkt, err := proto.Marshal(&pub)
	if err != nil {
		return 0, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_PUBLISH, RemainingLength: uint32(len(pkt))}
	buf := fh.pack()
	buf.Write(pkt)
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

// WriteTo writes the encoded Packet to the underlying writer.
func (p *Puback) WriteTo(w io.Writer) (int64, error) {
	puback := pbx.Puback{
		MessageID: uint32(p.MessageID),
	}
	pkt, err := proto.Marshal(&puback)
	if err != nil {
		return 0, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_PUBACK, RemainingLength: uint32(len(pkt))}
	buf := fh.pack()
	buf.Write(pkt)
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

// WriteTo writes the encoded Packet to the underlying writer.
func (p *Pubrec) WriteTo(w io.Writer) (int64, error) {
	pubrec := pbx.Pubrec{
		MessageID: uint32(p.MessageID),
	}
	pkt, err := proto.Marshal(&pubrec)
	if err != nil {
		return 0, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_PUBREC, RemainingLength: uint32(len(pkt))}
	buf := fh.pack()
	buf.Write(pkt)
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

// WriteTo writes the encoded Packet to the underlying writer.
func (p *Pubrel) WriteTo(w io.Writer) (int64, error) {
	pubrel := pbx.Pubrel{
		MessageID: uint32(p.MessageID),
	}
	pkt, err := proto.Marshal(&pubrel)
	if err != nil {
		return 0, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_PUBREL, RemainingLength: uint32(len(pkt))}
	buf := fh.pack()
	buf.Write(pkt)
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

// WriteTo writes the encoded Packet to the underlying writer.
func (p *Pubcomp) WriteTo(w io.Writer) (int64, error) {
	pubcomp := pbx.Pubcomp{
		MessageID: uint32(p.MessageID),
	}
	pkt, err := proto.Marshal(&pubcomp)
	if err != nil {
		return 0, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_PUBCOMP, RemainingLength: uint32(len(pkt))}
	buf := fh.pack()
	buf.Write(pkt)
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

func unpackPublish(data []byte) Packet {
	var pkt pbx.Publish
	proto.Unmarshal(data, &pkt)

	return &Publish{
		MessageID: uint16(pkt.MessageID),
		Topic:     pkt.Topic,
		Payload:   pkt.Payload,
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
	return &Pubrel{
		MessageID: uint16(pkt.MessageID),
	}
}

func unpackPubcomp(data []byte) Packet {
	var pkt pbx.Pubcomp
	proto.Unmarshal(data, &pkt)
	return &Pubcomp{
		MessageID: uint16(pkt.MessageID),
	}
}
