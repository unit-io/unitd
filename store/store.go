package store

import (
	"encoding/json"
	"errors"
	"time"

	adapter "github.com/frontnet/trace/db"
	"github.com/frontnet/trace/message"
	"github.com/frontnet/trace/pkg/log"

	"github.com/kelindar/binary"
)

var adp adapter.Adapter

type configType struct {
	// Configurations for individual adapters.
	Adapters map[string]json.RawMessage `json:"adapters"`
}

func openAdapter(jsonconf string) error {
	var config configType
	if err := json.Unmarshal([]byte(jsonconf), &config); err != nil {
		return errors.New("store: failed to parse config: " + err.Error() + "(" + jsonconf + ")")
	}

	if adp == nil {
		return errors.New("store: database adapter is missing")
	}

	if adp.IsOpen() {
		return errors.New("store: connection is already opened")
	}

	var adapterConfig string
	if config.Adapters != nil {
		adapterConfig = string(config.Adapters[adp.GetName()])
	}

	return adp.Open(adapterConfig)
}

// Open initializes the persistence system. Adapter holds a connection pool for a database instance.
// 	 name - name of the adapter rquested in the config file
//   jsonconf - configuration string
func Open(jsonconf string) error {
	if err := openAdapter(jsonconf); err != nil {
		return err
	}

	return nil
}

// Close terminates connection to persistent storage.
func Close() error {
	if adp.IsOpen() {
		return adp.Close()
	}

	return nil
}

// IsOpen checks if persistent storage connection has been initialized.
func IsOpen() bool {
	if adp != nil {
		return adp.IsOpen()
	}

	return false
}

// GetAdapterName returns the name of the current adater.
func GetAdapterName() string {
	if adp != nil {
		return adp.GetName()
	}

	return ""
}

// InitDb open the db connection. If jsconf is nil it will assume that the connection is already open.
// If it's non-nil, it will use the config string to open the DB connection first.
func InitDb(jsonconf string, reset bool) error {
	if !IsOpen() {
		if err := openAdapter(jsonconf); err != nil {
			return err
		}
	}
	panic("store: Init DB error")
}

// RegisterAdapter makes a persistence adapter available.
// If Register is called twice or if the adapter is nil, it panics.
func RegisterAdapter(name string, a adapter.Adapter) {
	if a == nil {
		panic("store: Register adapter is nil")
	}

	if adp != nil {
		panic("store: adapter '" + adp.GetName() + "' is already registered")
	}

	adp = a
}

// MessageStore is a Message struct to hold methods for persistence mapping for the Message object.
type MessageStore struct{}

// Message is the ancor for storing/retrieving Message objects
var Message MessageStore

func (m *MessageStore) Store(msg *message.Message) error {
	key := msg.ID
	TTL := time.Duration(msg.TTL)
	val, err := binary.Marshal(msg)
	if err != nil {
		log.Error("adapter.Store", "msg marshal error: "+err.Error())
		return err
	}

	return adp.StoreWithTTL(key, val, TTL)
}

func (m *MessageStore) Query(ssid message.Ssid, from, until time.Time, limit int) ([]message.Message, error) {
	prefix := message.GenPrefix(ssid, until.Unix())
	cutoff := from.Unix()

	return adp.Query(prefix, ssid, cutoff, limit)
}
