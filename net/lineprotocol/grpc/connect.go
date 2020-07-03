package grpc

import (
	"io"

	"github.com/golang/protobuf/proto"
	lp "github.com/unit-io/unitd/net/lineprotocol"
	pbx "github.com/unit-io/unitd/proto"
)

type (
	Connect    lp.Connect
	Connack    lp.Connack
	Pingreq    lp.Pingreq
	Pingresp   lp.Pingresp
	Disconnect lp.Disconnect
)

// WriteTo writes the encoded message to the underlying writer.
func (c *Connect) WriteTo(w io.Writer) (int64, error) {
	conn := pbx.Conn{
		ProtoName:     c.ProtoName,
		Version:       uint32(c.Version),
		UsernameFlag:  c.UsernameFlag,
		PasswordFlag:  c.PasswordFlag,
		CleanSessFlag: c.CleanSessFlag,
		KeepAlive:     uint32(c.KeepAlive),
		ClientID:      c.ClientID,
		Username:      c.Username,
		Password:      c.Password,
	}

	pkt, err := proto.Marshal(&conn)
	if err != nil {
		return 0, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_CONNECT, RemainingLength: uint32(len(pkt))}
	buf := fh.pack()
	buf.Write(pkt)
	return buf.WriteTo(w)
}

// Type returns the packet type.
func (c *Connect) Type() uint8 {
	return uint8(pbx.MessageType_CONNECT)
}

// String returns the name of operation.
func (c *Connect) String() string {
	return "connect"
}

// Encode writes the encoded message to the buffer.
func (c *Connack) WriteTo(w io.Writer) (int64, error) {
	connack := pbx.Connack{
		ReturnCode: uint32(c.ReturnCode),
		ConnID:     c.ConnID,
	}
	pkt, err := proto.Marshal(&connack)
	if err != nil {
		return 0, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_CONNACK, RemainingLength: uint32(len(pkt))}
	buf := fh.pack()
	buf.Write(pkt)
	return buf.WriteTo(w)
}

// Type returns the packet type.
func (c *Connack) Type() uint8 {
	return uint8(pbx.MessageType_CONNACK)
}

// String returns the name of operation.
func (c *Connack) String() string {
	return "connack"
}

// WriteTo writes the encoded packet to the underlying writer.
func (p *Pingreq) WriteTo(w io.Writer) (int64, error) {
	pingreq := pbx.Pingreq{}
	pkt, err := proto.Marshal(&pingreq)
	if err != nil {
		return 0, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_PINGREQ, RemainingLength: uint32(len(pkt))}
	buf := fh.pack()
	buf.Write(pkt)
	return buf.WriteTo(w)

}

// Type returns the packet type.
func (p *Pingreq) Type() uint8 {
	return uint8(pbx.MessageType_PINGREQ)
}

// String returns the name of operation.
func (p *Pingreq) String() string {
	return "pingreq"
}

// WriteTo writes the encoded packet to the underlying writer.
func (p *Pingresp) WriteTo(w io.Writer) (int64, error) {
	pingresp := pbx.Pingresp{}
	pkt, err := proto.Marshal(&pingresp)
	if err != nil {
		return 0, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_PINGRESP, RemainingLength: uint32(len(pkt))}
	buf := fh.pack()
	buf.Write(pkt)
	return buf.WriteTo(w)
}

// Type returns the packet type.
func (p *Pingresp) Type() uint8 {
	return uint8(pbx.MessageType_PINGRESP)
}

// String returns the name of operation.
func (p *Pingresp) String() string {
	return "pingresp"
}

// WriteTo writes the encoded packet to the underlying writer.
func (d *Disconnect) WriteTo(w io.Writer) (int64, error) {
	disc := pbx.Disconnect{}
	pkt, err := proto.Marshal(&disc)
	if err != nil {
		return 0, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_DISCONNECT, RemainingLength: uint32(len(pkt))}
	buf := fh.pack()
	buf.Write(pkt)
	return buf.WriteTo(w)
}

// Type returns the packet type.
func (d *Disconnect) Type() uint8 {
	return uint8(pbx.MessageType_DISCONNECT)
}

// String returns the name of operation.
func (d *Disconnect) String() string {
	return "disconnect"
}

func unpackConnect(data []byte) Packet {
	var pkt pbx.Conn
	proto.Unmarshal(data, &pkt)

	connect := &Connect{
		ProtoName:     pkt.ProtoName,
		Version:       uint8(pkt.Version),
		KeepAlive:     uint16(pkt.KeepAlive),
		ClientID:      pkt.ClientID,
		InsecureFlag:  pkt.InsecureFlag,
		UsernameFlag:  pkt.UsernameFlag,
		PasswordFlag:  pkt.PasswordFlag,
		CleanSessFlag: pkt.CleanSessFlag,
	}

	if connect.UsernameFlag {
		connect.Username = pkt.Username
	}

	if connect.PasswordFlag {
		connect.Password = pkt.Password
	}
	return connect
}

func unpackConnack(data []byte) Packet {
	var pkt pbx.Connack
	proto.Unmarshal(data, &pkt)

	return &Connack{
		ReturnCode: uint8(pkt.ReturnCode),
	}
}

func unpackPingreq(data []byte) Packet {
	return &Pingreq{}
}

func unpackPingresp(data []byte) Packet {
	return &Pingresp{}
}

func unpackDisconnect(data []byte) Packet {
	return &Disconnect{}
}
