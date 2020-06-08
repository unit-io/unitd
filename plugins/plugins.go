package plugins

import "fmt"

//Packet is the interface all our packets will be implementing
type Packet interface {
	fmt.Stringer

	Type() uint8
	Encode() []byte
}
