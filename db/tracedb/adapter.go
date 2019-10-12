package adapter

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/kelindar/binary"
	"github.com/saffat-in/trace/message"
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

// Store appends the messages to the store.
func (a *adapter) StoreWithTTL(key, val []byte, TTL time.Duration) error {
	// Start the transaction.
	// return a.db.PutWithTTL(key, val, TTL)
	return a.db.Batch(func(b *tracedb.Batch) error {
		b.Put(key, val)
		err := b.Write()
		return err
	})
}

// Query performs a query and attempts to fetch last n messages where
// n is specified by limit argument. From and until times can also be specified
// for time-series retrieval.
func (a *adapter) Query(prefix []byte, ssid message.Ssid, cutoff int64, limit int) (matches []message.Message, err error) {
	if limit == 0 {
		limit = maxResults // Maximum number of records to return
	}

	// Iterating over key/value pairs.
	it, err := a.db.Items(&tracedb.Query{Topic: []byte("dev18?last=3m")})

	// Seek the prefix and check the key so we can quickly exit the iteration.
	for it.First(); it.Valid() && message.ID(it.Item().Key()).EvalPrefix(ssid, cutoff) && len(matches) < maxResults && len(matches) < limit; it.Next() {
		//for it.Seek(prefix); it.Valid(); it.Next() {
		var msg message.Message
		err = binary.Unmarshal(it.Item().Value(), &msg)
		if err != nil {
			log.Error("adapter.Query", "unable to unmarshal value: "+err.Error())
			return nil, err
		}
		matches = append(matches, msg)

		if err != nil {
			log.Error("adapter.Query", "unable to query db: "+err.Error())
			return nil, err
		}
	}
	return matches, nil
}

func init() {
	store.RegisterAdapter(adapterName, &adapter{})
}
