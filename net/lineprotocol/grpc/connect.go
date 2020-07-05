package grpc

import (
	"bytes"
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

func (c *Connect) encode() (bytes.Buffer, error) {
	var buf bytes.Buffer
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
		return buf, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_CONNECT, RemainingLength: uint32(len(pkt))}
	buf = fh.pack()
	_, err = buf.Write(pkt)
	return buf, err
}

// Encode encodes message into binary data
func (c *Connect) Encode() []byte {
	buf, err := c.encode()
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

// WriteTo writes the encoded message to the underlying writer.
func (c *Connect) WriteTo(w io.Writer) (int64, error) {
	buf, err := c.encode()
	if err != nil {
		return 0, err
	}
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

// Info returns Qos and MessageID of this packet.
func (c *Connect) Info() lp.Info {
	return lp.Info{Qos: 0, MessageID: 0}
}

func (c *Connack) encode() (bytes.Buffer, error) {
	var buf bytes.Buffer
	connack := pbx.Connack{
		ReturnCode: uint32(c.ReturnCode),
		ConnID:     c.ConnID,
	}
	pkt, err := proto.Marshal(&connack)
	if err != nil {
		return buf, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_CONNACK, RemainingLength: uint32(len(pkt))}
	buf = fh.pack()
	_, err = buf.Write(pkt)
	return buf, err
}

// Encode encodes message into binary data
func (c *Connack) Encode() []byte {
	buf, err := c.encode()
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

// WriteTo writes the encoded message to the buffer.
func (c *Connack) WriteTo(w io.Writer) (int64, error) {
	buf, err := c.encode()
	if err != nil {
		return 0, err
	}
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

// Info returns Qos and MessageID of this packet.
func (c *Connack) Info() lp.Info {
	return lp.Info{Qos: 0, MessageID: 0}
}

func (p *Pingreq) encode() (bytes.Buffer, error) {
	var buf bytes.Buffer
	pingreq := pbx.Pingreq{}
	pkt, err := proto.Marshal(&pingreq)
	if err != nil {
		return buf, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_PINGREQ, RemainingLength: uint32(len(pkt))}
	buf = fh.pack()
	_, err = buf.Write(pkt)
	return buf, err
}

// Encode encodes message into binary data
func (p *Pingreq) Encode() []byte {
	buf, err := p.encode()
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

// WriteTo writes the encoded packet to the underlying writer.
func (p *Pingreq) WriteTo(w io.Writer) (int64, error) {
	buf, err := p.encode()
	if err != nil {
		return 0, err
	}
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

// Info returns Qos and MessageID of this packet.
func (p *Pingreq) Info() lp.Info {
	return lp.Info{Qos: 0, MessageID: 0}
}

func (p *Pingresp) encode() (bytes.Buffer, error) {
	var buf bytes.Buffer
	pingresp := pbx.Pingresp{}
	pkt, err := proto.Marshal(&pingresp)
	if err != nil {
		return buf, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_PINGRESP, RemainingLength: uint32(len(pkt))}
	buf = fh.pack()
	_, err = buf.Write(pkt)
	return buf, err
}

// Encode encodes message into binary data
func (p *Pingresp) Encode() []byte {
	buf, err := p.encode()
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

// WriteTo writes the encoded packet to the underlying writer.
func (p *Pingresp) WriteTo(w io.Writer) (int64, error) {
	buf, err := p.encode()
	if err != nil {
		return 0, err
	}
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

// Info returns Qos and MessageID of this packet.
func (p *Pingresp) Info() lp.Info {
	return lp.Info{Qos: 0, MessageID: 0}
}

func (d *Disconnect) encode() (bytes.Buffer, error) {
	var buf bytes.Buffer
	disc := pbx.Disconnect{}
	pkt, err := proto.Marshal(&disc)
	if err != nil {
		return buf, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_DISCONNECT, RemainingLength: uint32(len(pkt))}
	buf = fh.pack()
	_, err = buf.Write(pkt)
	return buf, err
}

// Encode encodes message into binary data
func (d *Disconnect) Encode() []byte {
	buf, err := d.encode()
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

// WriteTo writes the encoded packet to the underlying writer.
func (d *Disconnect) WriteTo(w io.Writer) (int64, error) {
	buf, err := d.encode()
	if err != nil {
		return 0, err
	}
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

// Info returns Qos and MessageID of this packet.
func (d *Disconnect) Info() lp.Info {
	return lp.Info{Qos: 0, MessageID: 0}
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
