package adapter

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/saffat-in/trace/pkg/collection"
	"github.com/saffat-in/trace/pkg/log"
	"github.com/saffat-in/trace/store"
	"github.com/saffat-in/tracedb"
)

const (
	defaultDatabase = "trace"

	dbVersion = 2.0

	adapterName = "tracedb"
)

type configType struct {
	Dir      string `json:"dir,omitempty"`
	ValueDir string `json:"value_dir,omitempty"`
}

const (
	// Maximum number of records to return
	maxResults = 1024
	// Maximum TTL for message
	maxTTL = "24h"
)

// Store represents an SSD-optimized storage store.
type adapter struct {
	db      *tracedb.DB // The underlying database to store messages.
	version int
}

// Open initializes database connection
func (a *adapter) Open(jsonconfig string) error {
	if a.db != nil {
		return errors.New("tracedb adapter is already connected")
	}

	var err error
	var config configType

	if err = json.Unmarshal([]byte(jsonconfig), &config); err != nil {
		return errors.New("tracedb adapter failed to parse config: " + err.Error())
	}

	// Make sure we have a directory
	if err := os.MkdirAll(config.Dir, 0777); err != nil {
		log.Error("adapter.Open", "Unable to create db dir")
	}

	// Attempt to open the database
	a.db, err = tracedb.Open(config.Dir+"/"+defaultDatabase, nil)
	if err != nil {
		log.Error("adapter.Open", "Unable to open db")
		return err
	}
	return nil
}

// Close closes the underlying database connection
func (a *adapter) Close() error {
	var err error
	if a.db != nil {
		err = a.db.Close()
		a.db = nil
		a.version = -1
	}
	return err
}

// IsOpen returns true if connection to database has been established. It does not check if
// connection is actually live.
func (a *adapter) IsOpen() bool {
	return a.db != nil
}

// GetName returns string that adapter uses to register itself with store.
func (a *adapter) GetName() string {
	return adapterName
}

// Put appends the messages to the store.
func (a *adapter) Put(contract uint32, topic, payload []byte) error {
	err := a.db.PutEntry(&tracedb.Entry{
		Topic:    topic,
		Payload:  payload,
		Contract: contract,
	})
	return err
}

// PutWithID appends the messages to the store using a pre generated messageId.
func (a *adapter) PutWithID(contract uint32, topic, messageId, payload []byte) error {
	err := a.db.PutEntry(&tracedb.Entry{
		ID:       messageId,
		Topic:    topic,
		Payload:  payload,
		Contract: contract,
	})
	return err
}

// Get performs a query and attempts to fetch last n messages where
// n is specified by limit argument. From and until times can also be specified
// for time-series retrieval.
func (a *adapter) Get(contract uint32, topic []byte, limit int) (matches []collection.Payload, err error) {
	// Iterating over key/value pairs.
	it, err := a.db.Items(&tracedb.Query{Topic: topic, Contract: contract, Limit: uint32(limit)})
	if err != nil {
		return nil, err
	}
	// Seek the prefix and check the key so we can quickly exit the iteration.
	for it.First(); it.Valid(); it.Next() {
		if err := it.Error(); err != nil {
			log.Error("adapter.Query", "unable to query db: "+err.Error())
			return nil, err
		}
		matches = append(matches, it.Item().Value())
	}
	return matches, nil
}

// GenID generates a messageId.
func (a *adapter) GenID(contract uint32, topic, payload []byte) ([]byte, error) {
	id := a.db.GenID()
	if id == nil {
		return nil, errors.New("Key is empty.")
	}
	return id, nil
}

// Put appends the messages to the store.
func (a *adapter) Delete(contract uint32, topic, messageId []byte) error {
	err := a.db.DeleteEntry(&tracedb.Entry{
		ID:       messageId,
		Topic:    topic,
		Contract: contract,
	})
	return err
}

func init() {
	store.RegisterAdapter(adapterName, &adapter{})
}
