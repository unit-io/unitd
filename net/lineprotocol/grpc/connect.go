package grpc

import (
	"bytes"

	"github.com/golang/protobuf/proto"
	lp "github.com/unit-io/unitd/net/lineprotocol"
	pbx "github.com/unit-io/unitd/proto"
)

func encodeConnect(c lp.Connect) (bytes.Buffer, error) {
	var msg bytes.Buffer
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
		return msg, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_CONNECT, RemainingLength: uint32(len(pkt))}
	msg = fh.pack()
	_, err = msg.Write(pkt)
	return msg, err
}

func encodeConnack(c lp.Connack) (bytes.Buffer, error) {
	var msg bytes.Buffer
	connack := pbx.Connack{
		ReturnCode: uint32(c.ReturnCode),
		ConnID:     c.ConnID,
	}
	pkt, err := proto.Marshal(&connack)
	if err != nil {
		return msg, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_CONNACK, RemainingLength: uint32(len(pkt))}
	msg = fh.pack()
	_, err = msg.Write(pkt)
	return msg, err
}

func encodePingreq(p lp.Pingreq) (bytes.Buffer, error) {
	var msg bytes.Buffer
	pingreq := pbx.Pingreq{}
	pkt, err := proto.Marshal(&pingreq)
	if err != nil {
		return msg, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_PINGREQ, RemainingLength: uint32(len(pkt))}
	msg = fh.pack()
	_, err = msg.Write(pkt)
	return msg, err
}

func encodePingresp(p lp.Pingresp) (bytes.Buffer, error) {
	var msg bytes.Buffer
	pingresp := pbx.Pingresp{}
	pkt, err := proto.Marshal(&pingresp)
	if err != nil {
		return msg, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_PINGRESP, RemainingLength: uint32(len(pkt))}
	msg = fh.pack()
	_, err = msg.Write(pkt)
	return msg, err
}

func encodeDisconnect(d lp.Disconnect) (bytes.Buffer, error) {
	var msg bytes.Buffer
	disc := pbx.Disconnect{}
	pkt, err := proto.Marshal(&disc)
	if err != nil {
		return msg, err
	}
	fh := FixedHeader{MessageType: pbx.MessageType_DISCONNECT, RemainingLength: uint32(len(pkt))}
	msg = fh.pack()
	_, err = msg.Write(pkt)
	return msg, err
}

func unpackConnect(data []byte) lp.Packet {
	var pkt pbx.Conn
	proto.Unmarshal(data, &pkt)

	connect := &lp.Connect{
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

func unpackConnack(data []byte) lp.Packet {
	var pkt pbx.Connack
	proto.Unmarshal(data, &pkt)

	return &lp.Connack{
		ReturnCode: uint8(pkt.ReturnCode),
	}
}

func unpackPingreq(data []byte) lp.Packet {
	return &lp.Pingreq{}
}

func unpackPingresp(data []byte) lp.Packet {
	return &lp.Pingresp{}
}

func unpackDisconnect(data []byte) lp.Packet {
	return &lp.Disconnect{}
}
