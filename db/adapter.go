package adapter

import (
	"errors"
	"time"

	"github.com/tracedb/trace/message"
)

var (
	errNotFound = errors.New("no messages were found")
)

// Storage represents a message storage contract that message storage provides
// must fulfill.
type Adapter interface {
	// General

	// Open and configure the adapter
	Open(config string) error
	// Close the adapter
	Close() error
	// IsOpen checks if the adapter is ready for use
	IsOpen() bool
	// // CheckDbVersion checks if the actual database version matches adapter version.
	// CheckDbVersion() error
	// GetName returns the name of the adapter
	GetName() string

	// Store is used to store a message, the SSID provided must be a full SSID
	// SSID, where first element should be a contract ID. The time resolution
	// for TTL will be in seconds. The function is executed synchronously and
	// it returns an error if some error was encountered during storage.
	StoreWithTTL(key, val []byte, TTL time.Duration) error

	// Query performs a query and attempts to fetch last n messages where
	// n is specified by limit argument. From and until times can also be specified
	// for time-series retrieval.
	Query(prefix []byte, ssid message.Ssid, cutoff int64, limit int) ([]message.Message, error)
}
