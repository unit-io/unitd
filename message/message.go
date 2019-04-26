package message

import (
	"sync/atomic"
	"math"

	"github.com/kelindar/binary"
	"github.com/tracedb/trace/pkg/uid"
)

const (
	CONNECT = uint8(iota + 1)
	PUBLISH
	SUBSCRIBE
	UNSUBSCRIBE
	PINGREQ
	PINGRESP
	DISCONNECT
	
	fixed = 16
)

// ID represents a message ID encoded at 128bit and lexigraphically sortable
type ID []byte

// NewID creates a new message identifier for the current time.
func NewID(ssid Ssid) ID {
	id := make(ID, len(ssid)*4+fixed)

	binary.BigEndian.PutUint32(id[0:4], ssid[0]^ssid[1])
	binary.BigEndian.PutUint32(id[4:8], uid.NewApoch())
	binary.BigEndian.PutUint32(id[8:12], math.MaxUint32-atomic.AddUint32(&uid.Next, 1)) // Reverse order
	binary.BigEndian.PutUint32(id[12:16], uid.NewUnique())
	for i, v := range ssid {
		binary.BigEndian.PutUint32(id[fixed+i*4:fixed+4+i*4], v)
	}

	return id
}

// Message represents a message which has to be forwarded or stored.
type Message struct {
	ID      ID     `json:"id,omitempty"`   // The ID of the message
	Topic   []byte `json:"chan,omitempty"` // The topic of the message
	Payload []byte `json:"data,omitempty"` // The payload of the message
	TTL     int64 `json:"ttl,omitempty"`  // The time-to-live of the message
}

// New creates a new message structure from the provided SSID, topic and payload.
func New(ssid Ssid, topic, payload []byte) *Message {
	return &Message{
		ID:      NewID(ssid),
		Topic:   topic,
		Payload: payload,
	}
}

// Size returns the byte size of the message.
func (m *Message) Size() int64 {
	return int64(len(m.Payload))
}

// GenPrefix generates a new message identifier only containing the prefix.
func GenPrefix(ssid Ssid, from int64) ID {
	id := make(ID, 8)
	if len(ssid) < 2 {
		return id
	}

	binary.BigEndian.PutUint32(id[0:4], ssid[0]^ssid[1])
	binary.BigEndian.PutUint32(id[4:8], math.MaxUint32-uint32(from-uid.Offset))
	
	return id
}

// Time gets the time of the key, adjusted.
func (id ID) Time() int64 {
	return int64(math.MaxUint32-binary.BigEndian.Uint32(id[4:8])) + uid.Offset
}

// EvalPrefix matches the prefix with the cutoff time.
func (id ID) EvalPrefix(ssid Ssid, cutoff int64) bool {
	return (binary.BigEndian.Uint32(id[0:4]) == ssid[0]^ssid[1]) && id.Time() >= cutoff
}
