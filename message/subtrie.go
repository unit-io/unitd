package message

import (
	"sync"
)

const nul = 0x0

type key struct {
	query       uint32
	wildchars uint8
}

type part struct {
	depth    uint8
	subs     Subscribers
	parent   *part
	children map[key]*part
}

// Trie represents an efficient collection of subscriptions with lookup capability.
type PartTrie struct {
	root *part // The root node of the tree.
}

// NewTrie creates a new matcher for the subscriptions.
func NewPartTrie() *PartTrie {
	return &PartTrie{
		root: &part{
			subs:     Subscribers{},
			children: make(map[key]*part),
		},
	}
}

type Subscriptions struct {
	sync.RWMutex
	parttrie *PartTrie
	count    int // Number of subscriptions in the trie.
}

// Creates a subscriptions with an initialized trie.
func NewSubscriptions() *Subscriptions {
	return &Subscriptions{
		parttrie: NewPartTrie(),
	}
}

// Count returns the number of subscriptions.
func (s *Subscriptions) Count() int {
	s.RLock()
	defer s.RUnlock()
	return s.count
}

// Subscribe adds the Subscriber to the topic and returns a Subscription.
func (s *Subscriptions) Subscribe(parts []Part, depth uint8, subscriber Subscriber) error {
	s.Lock()
	defer s.Unlock()
	curr := s.parttrie.root
	for _, p := range parts {
		k := key{
			query:       p.Query,
			wildchars: p.Wildchars,
		}
		child, ok := curr.children[k]
		if !ok {
			child = &part{
				subs:     Subscribers{},
				parent:   curr,
				children: make(map[key]*part),
			}
			curr.children[k] = child
		}
		curr = child
	}
	if ok := curr.subs.AddUnique(subscriber); ok {
		curr.depth = depth
		s.count++
	}

	return nil
}

// Unsubscribe remove the subscription for the topic.
func (s *Subscriptions) Unsubscribe(parts []Part, subscriber Subscriber) error {
	s.Lock()
	defer s.Unlock()
	curr := s.parttrie.root

	for _, part := range parts {
		k := key{
			query:       part.Query,
			wildchars: part.Wildchars,
		}
		child, ok := curr.children[k]
		if !ok {
			// Subscription doesn't exist.
			return nil
		}
		curr = child
	}
	// Remove the subscriber and decrement the counter
	if ok := curr.subs.Remove(subscriber); ok {
		s.count--
	}
	return nil
}

// Lookup returns the Subscribers for the given topic.
func (s *Subscriptions) Lookup(query Ssid) (subs Subscribers) {
	s.RLock()
	defer s.RUnlock()

	s.lookup(query, uint8(len(query)-1), &subs, s.parttrie.root)
	return
}

func (s *Subscriptions) lookup(query Ssid, depth uint8, subs *Subscribers, part *part) {
	// Add subscribers from the current branch
	for _, s := range part.subs {
		if part.depth == depth || (part.depth >= 23 && depth > part.depth-23) {
			subs.AddUnique(s)
		}
	}

	// If we're not yet done, continue
	if len(query) > 0 {
		// Go through the exact match branch
		for k, p := range part.children {
			if k.query == query[0] && uint8(len(query)) >= k.wildchars+1 {
				s.lookup(query[k.wildchars+1:], depth, subs, p)
			}
		}
	}
}
