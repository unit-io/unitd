package mqtt

import (
	"bytes"
)

// Connect represents an MQTT connect packet.
type Connect struct {
	ProtoName      []byte
	Version        uint8
	UsernameFlag   bool
	PasswordFlag   bool
	WillRetainFlag bool
	WillQOS        uint8
	WillFlag       bool
	CleanSeshFlag  bool
	KeepAlive      uint16
	ClientID       []byte
	WillTopic      []byte
	WillMessage    []byte
	Username       []byte
	Password       []byte
}

// Connack represents an MQTT connack packet.
// 0x00 connection accepted
// 0x01 refused: unacceptable proto version
// 0x02 refused: identifier rejected
// 0x03 refused server unavailiable
// 0x04 bad user or password
// 0x05 not authorized
type Connack struct {
	ReturnCode uint8
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

// EncodeTo writes the encoded message to the underlying writer.
func (c *Connect) Encode() []byte {
	var buf bytes.Buffer

	//write some buffer space in the beginning for the maximum number of bytes static header + legnth encoding can take
	//buf.Write(reserveForHeader)

	// pack the proto name and version
	buf.Write(c.ProtoName)
	buf.WriteByte(c.Version)

	// pack the flags
	var flagByte byte
	flagByte |= byte(boolToUInt8(c.UsernameFlag)) << 7
	flagByte |= byte(boolToUInt8(c.PasswordFlag)) << 6
	flagByte |= byte(boolToUInt8(c.WillRetainFlag)) << 5
	flagByte |= byte(c.WillQOS) << 3
	flagByte |= byte(boolToUInt8(c.WillFlag)) << 2
	flagByte |= byte(boolToUInt8(c.CleanSeshFlag)) << 1
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
	packet := encodeParts(CONNECT, buf.Len(), nil)
	packet.Write(buf.Bytes())
	return packet.Bytes()
}

// Type returns the MQTT packet type.
func (c *Connect) Type() uint8 {
	return CONNECT
}

// String returns the name of mqtt operation.
func (c *Connect) String() string {
	return "connect"
}

// EncodeTo writes the encoded packet to the underlying writer.
func (c *Connack) Encode() []byte {
	var buf bytes.Buffer

	//write padding
	//buf.Write(reserveForHeader)
	buf.WriteByte(byte(0))
	buf.WriteByte(byte(c.ReturnCode))

	// Write to the underlying buffer
	packet := encodeParts(CONNACK, 2, nil)
	packet.Write(buf.Bytes())
	return packet.Bytes()
}

// Type returns the MQTT packet type.
func (c *Connack) Type() uint8 {
	return CONNACK
}

// String returns the name of mqtt operation.
func (c *Connack) String() string {
	return "connack"
}

// EncodeTo writes the encoded packet to the underlying writer.
func (p *Pingreq) Encode() []byte {
	return []byte{0xc0, 0x0}
}

// Type returns the MQTT packet type.
func (p *Pingreq) Type() uint8 {
	return PINGREQ
}

// String returns the name of mqtt operation.
func (p *Pingreq) String() string {
	return "pingreq"
}

// EncodeTo writes the encoded packet to the underlying writer.
func (p *Pingresp) Encode() []byte {
	return []byte{0xd0, 0x0}
}

// Type returns the MQTT packet type.
func (p *Pingresp) Type() uint8 {
	return PINGRESP
}

// String returns the name of mqtt operation.
func (p *Pingresp) String() string {
	return "pingresp"
}

// EncodeTo writes the encoded packet to the underlying writer.
func (d *Disconnect) Encode() []byte {
	return []byte{0xe0, 0x0}
}

// Type returns the MQTT packet type.
func (d *Disconnect) Type() uint8 {
	return DISCONNECT
}

// String returns the name of mqtt operation.
func (d *Disconnect) String() string {
	return "disconnect"
}

func unpackConnect(data []byte, hdr *FixedHeader) Packet {
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
		CleanSeshFlag:  flags&(1<<1) > 0,
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

func unpackConnack(data []byte, hdr *FixedHeader) Packet {
	//first byte is weird in connack
	bookmark := uint32(1)
	retcode := data[bookmark]

	return &Connack{
		ReturnCode: retcode,
	}
}

func unpackPingreq(data []byte, hdr *FixedHeader) Packet {
	return &Pingreq{}
}

func unpackPingresp(data []byte, hdr *FixedHeader) Packet {
	return &Pingresp{}
}

func unpackDisconnect(data []byte, hdr *FixedHeader) Packet {
	return &Disconnect{}
}
