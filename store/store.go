package store

import (
	"encoding/binary"
	"encoding/json"
	"errors"

	adapter "github.com/saffat-in/trace/db"
	"github.com/saffat-in/trace/message"
	"github.com/saffat-in/trace/pkg/uid"
)

const (
	// Maximum number of records to return
	maxResults         = 1024
	connStoreId uint32 = 4105991048 // hash("connectionstore")
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

func (m *MessageStore) Put(contract uint32, topic, payload []byte) error {
	return adp.Put(contract, topic, payload)
}

func (m *MessageStore) Get(contract uint32, topic []byte) (matches []message.Message, err error) {
	resp, err := adp.Get(contract, topic, 0)
	for _, payload := range resp {
		msg := message.Message{
			Topic:   topic,
			Payload: payload,
		}
		matches = append(matches, msg)
	}

	return matches, err
}

// ConnectionStore is a Conection struct to hold methods for persistence mapping for the Connect LId.
// Note, do not use same contract as messagestore
type ConnectionStore struct{}

// Message is the ancor for storing/retrieving Message objects
var Connection ConnectionStore

func (c *ConnectionStore) Put(contract uint32, topic []byte, messageId []byte, connId uid.LID) error {
	payload := make([]byte, 8)
	binary.LittleEndian.PutUint64(payload[:8], uint64(connId))
	return adp.PutWithID(contract^connStoreId, topic, messageId, payload)
}

func (c *ConnectionStore) Get(contract uint32, topic []byte) (matches []uid.LID, err error) {
	resp, err := adp.Get(contract^connStoreId, topic, maxResults)
	for _, payload := range resp {
		matches = append(matches, uid.LID(binary.LittleEndian.Uint64(payload[:])))
	}

	return matches, err
}

func (c *ConnectionStore) NewID(contract uint32, topic []byte, connId uid.LID) ([]byte, error) {
	payload := make([]byte, 8)
	binary.LittleEndian.PutUint64(payload[:8], uint64(connId))
	return adp.NewID(contract^connStoreId, topic, payload)
}

func (c *ConnectionStore) Delete(contract uint32, topic []byte, messageId []byte) error {
	return adp.Delete(contract^connStoreId, topic, messageId)
}
