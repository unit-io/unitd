package message

import (
	"unsafe"
	"sync"
)

// Various constant parts of the SSID.
const (
	system   = uint32(0)
	Contract = uint32(3376684800)
	wildcard = uint32(857445537)
)

// Ssid represents a subscription ID which contains a contract and a list of hashes
// for various parts of the topic.
type Ssid []uint32

// NewSsid creates a new SSID.
func NewSsid(parts []Part) Ssid{
	ssid := make([]uint32, 0, len(parts))
	for _, part := range parts {
		ssid = append(ssid, part.Query)
	}
	return ssid
}

// AddContract adds contract to the parts.
func AddContract(contract uint32, c *Topic) {
	part := Part{
		Wildchars: 0,
		Query:     contract,
	}
	if c.Parts[0].Query == wildcard {
		c.Parts[0].Query = contract
	} else {
		parts := []Part{part}
		c.Parts = append(parts, c.Parts...)
	}
}

// unsafeToString is used to convert a slice
// of bytes to a string without incurring overhead.
func unsafeToString(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}

// ------------------------------------------------------------------------------------

// SubscriberType represents a type of subscriber
type SubscriberType uint8

type TopicAnyCount uint8

// Subscriber types
const (
	SubscriberDirect = SubscriberType(iota)
	SubscriberRemote
)

// Subscriber is a value associated with a subscription.
type Subscriber interface {
	ID() string
	Type() SubscriberType
	SendMessage(*Message) bool
}

// ------------------------------------------------------------------------------------

// Subscribers represents a subscriber set which can contain only unique values.
type Subscribers []Subscriber

// AddUnique adds a subscriber to the set.
func (s *Subscribers) AddUnique(value Subscriber) (added bool) {
	if s.Contains(value) == false {
		*s = append(*s, value)
		added = true
	}
	return
}

// Remove removes a subscriber from the set.
func (s *Subscribers) Remove(value Subscriber) (removed bool) {
	for i, v := range *s {
		if v == value {
			a := *s
			a[i] = a[len(a)-1]
			a[len(a)-1] = nil
			a = a[:len(a)-1]
			*s = a
			removed = true
			return
		}
	}
	return
}

// Contains checks whether a subscriber is in the set.
func (s *Subscribers) Contains(value Subscriber) bool {
	for _, v := range *s {
		if v == value {
			return true
		}
	}
	return false
}

// Counters represents a subscription counting map.
type Counters struct {
	sync.Mutex
	stats map[string][][]Part
}

// NewCounters creates a new container.
func NewCounters() *Counters {
	return &Counters{
		stats: make(map[string][][]Part),
	}
}

// Increment adds the subscription to the stats.
func (s *Counters) Increment(parts []Part, key string) (first bool) {
	s.Lock()
	defer s.Unlock()

	if _, exists := s.stats[key]; exists {
		return len(s.stats[key]) == 1
	}
	s.stats[key] = append(s.stats[key], parts)
	return len(s.stats[key]) == 1
}

// Decrement remove a subscription from the stats.
func (s *Counters) Decrement(parts []Part, key string) (last bool) {
	s.Lock()
	defer s.Unlock()

	l := len(s.stats[key])
	if l > 0 {
		s.stats[key] = s.stats[key][:l-1]
	}

	// Remove if there's no subscribers left
	if len(s.stats[key]) <= 0 {
		delete(s.stats, key)
		return true
	}

	return false
}

// Get gets subscription from the stats.
func (s *Counters) Exist(key string) (ok bool) {
	s.Lock()
	defer s.Unlock()

	if _, exists := s.stats[key]; exists {
		return true
	}
	return false
}

// All gets the all subscriptions from the stats.
func (s *Counters) All() [][]Part {
	s.Lock()
	defer s.Unlock()
	var parts [][]Part
	for k, _ := range s.stats {
		for _, part := range s.stats[k] {
			parts = append(parts, part)
		}
	}

	return parts
}
