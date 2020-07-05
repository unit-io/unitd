package mqtt

import (
	"bytes"
	"io"

	lp "github.com/unit-io/unitd/net/lineprotocol"
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

	//write some buffer space in the beginning for the maximum number of bytes static header + legnth encoding can take
	//buf.Write(reserveForHeader)

	// pack the lp name and version
	buf.Write(c.ProtoName)
	buf.WriteByte(c.Version)

	// pack the flags
	var flagByte byte
	flagByte |= byte(boolToUInt8(c.UsernameFlag)) << 7
	flagByte |= byte(boolToUInt8(c.PasswordFlag)) << 6
	flagByte |= byte(boolToUInt8(c.WillRetainFlag)) << 5
	flagByte |= byte(c.WillQOS) << 3
	flagByte |= byte(boolToUInt8(c.WillFlag)) << 2
	flagByte |= byte(boolToUInt8(c.CleanSessFlag)) << 1
	buf.WriteByte(flagByte)

	buf.Write(encodeUint16(c.KeepAlive))
	buf.Write(c.ClientID)

	if c.WillFlag {
		buf.Write(c.WillTopic)
		buf.Write(c.WillMessage)
	}

	if c.UsernameFlag {
		buf.Write(c.Username)
	}

	if c.PasswordFlag {
		buf.Write(c.Password)
	}

	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.CONNECT, RemainingLength: buf.Len()}
	packet := fh.pack(nil)
	_, err := packet.Write(buf.Bytes())
	return packet, err
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

// Type returns the MQTT packet type.
func (c *Connect) Type() uint8 {
	return lp.CONNECT
}

// String returns the name of mqtt operation.
func (c *Connect) String() string {
	return "connect"
}

// Info returns Qos and MessageID of this packet.
func (c *Connect) Info() lp.Info {
	return lp.Info{Qos: 0, MessageID: 0}
}

func (c *Connack) encode() (bytes.Buffer, error) {
	var buf bytes.Buffer

	//buf.Write(reserveForHeader)
	buf.WriteByte(byte(0))
	buf.WriteByte(byte(c.ReturnCode))

	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.CONNACK, RemainingLength: 2}
	packet := fh.pack(nil)
	_, err := packet.Write(buf.Bytes())
	return packet, err
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

// Type returns the MQTT packet type.
func (c *Connack) Type() uint8 {
	return lp.CONNACK
}

// String returns the name of mqtt operation.
func (c *Connack) String() string {
	return "connack"
}

// Info returns Qos and MessageID of this packet.
func (c *Connack) Info() lp.Info {
	return lp.Info{Qos: 0, MessageID: 0}
}

// Encode encodes message into binary data
func (p *Pingreq) Encode() []byte {
	return []byte{0xc0, 0x0}
}

// WriteTo writes the encoded packet to the underlying writer.
func (p *Pingreq) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte{0xc0, 0x0})
	return int64(n), err
}

// Type returns the MQTT packet type.
func (p *Pingreq) Type() uint8 {
	return lp.PINGREQ
}

// String returns the name of mqtt operation.
func (p *Pingreq) String() string {
	return "pingreq"
}

// Info returns Qos and MessageID of this packet.
func (p *Pingreq) Info() lp.Info {
	return lp.Info{Qos: 0, MessageID: 0}
}

// Encode encodes message into binary data
func (p *Pingresp) Encode() []byte {
	return []byte{0xc0, 0x0}
}

// WriteTo writes the encoded packet to the underlying writer.
func (p *Pingresp) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte{0xc0, 0x0})
	return int64(n), err
}

// Type returns the MQTT packet type.
func (p *Pingresp) Type() uint8 {
	return lp.PINGRESP
}

// String returns the name of mqtt operation.
func (p *Pingresp) String() string {
	return "pingresp"
}

// Info returns Qos and MessageID of this packet.
func (p *Pingresp) Info() lp.Info {
	return lp.Info{Qos: 0, MessageID: 0}
}

// Encode encodes message into binary data
func (d *Disconnect) Encode() []byte {
	return []byte{0xc0, 0x0}
}

// WriteTo writes the encoded packet to the underlying writer.
func (d *Disconnect) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte{0xc0, 0x0})
	return int64(n), err
}

// Type returns the MQTT packet type.
func (d *Disconnect) Type() uint8 {
	return lp.DISCONNECT
}

// String returns the name of mqtt operation.
func (d *Disconnect) String() string {
	return "disconnect"
}

// Info returns Qos and MessageID of this packet.
func (d *Disconnect) Info() lp.Info {
	return lp.Info{Qos: 0, MessageID: 0}
}

func unpackConnect(data []byte, fh FixedHeader) Packet {
	//TODO: Decide how to recover rom invalid packets (offsets don't equal actual reading?)
	bookmark := uint32(0)

	protoname := readString(data, &bookmark)
	ver := uint8(data[bookmark])
	bookmark++
	flags := data[bookmark]
	bookmark++
	keepalive := readUint16(data, &bookmark)
	cliID := readString(data, &bookmark)
	connect := &Connect{
		ProtoName:      protoname,
		Version:        ver,
		KeepAlive:      keepalive,
		ClientID:       cliID,
		UsernameFlag:   flags&(1<<7) > 0,
		PasswordFlag:   flags&(1<<6) > 0,
		WillRetainFlag: flags&(1<<5) > 0,
		WillQOS:        (flags & (1 << 4)) + (flags & (1 << 3)),
		WillFlag:       flags&(1<<2) > 0,
		CleanSessFlag:  flags&(1<<1) > 0,
	}

	if connect.WillFlag {
		connect.WillTopic = readString(data, &bookmark)
		connect.WillMessage = readString(data, &bookmark)
	}

	if connect.UsernameFlag {
		connect.Username = readString(data, &bookmark)
	}

	if connect.PasswordFlag {
		connect.Password = readString(data, &bookmark)
	}
	return connect
}

func unpackConnack(data []byte, fh FixedHeader) Packet {
	//first byte is weird in connack
	bookmark := uint32(1)
	retcode := data[bookmark]

	return &Connack{
		ReturnCode: retcode,
	}
}

func unpackPingreq(data []byte, fh FixedHeader) Packet {
	return &Pingreq{}
}

func unpackPingresp(data []byte, fh FixedHeader) Packet {
	return &Pingresp{}
}

func unpackDisconnect(data []byte, fh FixedHeader) Packet {
	return &Disconnect{}
}
