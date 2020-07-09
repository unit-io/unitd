package mqtt

import (
	"bytes"

	lp "github.com/unit-io/unitd/net/lineprotocol"
)

func encodeConnect(c lp.Connect) (bytes.Buffer, error) {
	var msg bytes.Buffer

	// pack the lp name and version
	msg.Write(encodeBytes(c.ProtoName))
	msg.WriteByte(c.Version)

	// pack the flags
	var flagByte byte
	flagByte |= byte(boolToUInt8(c.UsernameFlag)) << 7
	flagByte |= byte(boolToUInt8(c.PasswordFlag)) << 6
	flagByte |= byte(boolToUInt8(c.WillRetainFlag)) << 5
	flagByte |= byte(c.WillQOS) << 3
	flagByte |= byte(boolToUInt8(c.WillFlag)) << 2
	flagByte |= byte(boolToUInt8(c.CleanSessFlag)) << 1
	msg.WriteByte(flagByte)

	msg.Write(encodeUint16(c.KeepAlive))
	msg.Write(encodeBytes(c.ClientID))

	if c.WillFlag {
		msg.Write(c.WillTopic)
		msg.Write(c.WillMessage)
	}

	if c.UsernameFlag {
		msg.Write(encodeBytes(c.Username))
	}

	if c.PasswordFlag {
		msg.Write(encodeBytes(c.Password))
	}

	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.CONNECT, RemainingLength: msg.Len()}
	packet := fh.pack(nil)
	_, err := packet.Write(msg.Bytes())
	return packet, err
}

func encodeConnack(c lp.Connack) (bytes.Buffer, error) {
	var msg bytes.Buffer

	//msg.Write(reserveForHeader)
	msg.WriteByte(byte(0))
	msg.WriteByte(byte(c.ReturnCode))

	// Write to the underlying buffer
	fh := FixedHeader{MessageType: lp.CONNACK, RemainingLength: 2}
	packet := fh.pack(nil)
	_, err := packet.Write(msg.Bytes())
	return packet, err
}

// Encode encodes message into binary data
func encodePingreq(p lp.Pingreq) (bytes.Buffer, error) {
	var msg bytes.Buffer
	_, err := msg.Write([]byte{0xc0, 0x0})
	return msg, err
}

// Encode encodes message into binary data
func encodePingresp(p lp.Pingresp) (bytes.Buffer, error) {
	var msg bytes.Buffer
	_, err := msg.Write([]byte{0xc0, 0x0})
	return msg, err
}

// Encode encodes message into binary data
func encodeDisconnect(d lp.Disconnect) (bytes.Buffer, error) {
	var msg bytes.Buffer
	_, err := msg.Write([]byte{0xc0, 0x0})
	return msg, err
}

func unpackConnect(data []byte, fh FixedHeader) lp.Packet {
	//TODO: Decide how to recover rom invalid packets (offsets don't equal actual reading?)
	bookmark := uint32(0)

	protoname := readString(data, &bookmark)
	ver := uint8(data[bookmark])
	bookmark++
	flags := data[bookmark]
	bookmark++
	keepalive := readUint16(data, &bookmark)
	cliID := readString(data, &bookmark)
	connect := &lp.Connect{
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

func unpackConnack(data []byte, fh FixedHeader) lp.Packet {
	//first byte is weird in connack
	bookmark := uint32(1)
	retcode := data[bookmark]

	return &lp.Connack{
		ReturnCode: retcode,
	}
}

func unpackPingreq(data []byte, fh FixedHeader) lp.Packet {
	return &lp.Pingreq{}
}

func unpackPingresp(data []byte, fh FixedHeader) lp.Packet {
	return &lp.Pingresp{}
}

func unpackDisconnect(data []byte, fh FixedHeader) lp.Packet {
	return &lp.Disconnect{}
}
