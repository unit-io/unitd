package adapter

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/saffat-in/trace/message"
	"github.com/saffat-in/trace/pkg/log"
	"github.com/saffat-in/trace/store"
	"github.com/kelindar/binary"
)

const (
	defaultDatabase = "trace"

	dbVersion = 2.0

	adapterName = "badger"
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
	db      *badger.DB // The underlying database to store messages.
	version int
}

// Open initializes database connection
func (a *adapter) Open(jsonconfig string) error {
	if a.db != nil {
		return errors.New("badger adapter is already connected")
	}

	var err error
	var config configType

	if err = json.Unmarshal([]byte(jsonconfig), &config); err != nil {
		return errors.New("badger adapter failed to parse config: " + err.Error())
	}

	// Make sure we have a directory
	if err := os.MkdirAll(config.Dir, 0777); err != nil {
		log.Error("adapter.Open", "Unable to create db dir")
	}

	// Attempt to open the database
	a.db, err = badger.Open(badger.DefaultOptions(config.Dir))
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
	tx := a.db.NewTransaction(true)
	defer tx.Discard()

	// Use msg ID as key
	e := badger.NewEntry(key, val).WithTTL(TTL)
	err := tx.SetEntry(e)
	if err != nil {
		log.Error("adapter.Store", "store error: "+err.Error())
		return err
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		log.Error("adapter.Store", "tx commit error: "+err.Error())
		return err
	}

	return nil
}

// Query performs a query and attempts to fetch last n messages where
// n is specified by limit argument. From and until times can also be specified
// for time-series retrieval.
func (a *adapter) Query(prefix []byte, ssid message.Ssid, cutoff int64, limit int) (matches []message.Message, err error) {
	if err = a.db.View(func(tx *badger.Txn) error {
		it := tx.NewIterator(badger.IteratorOptions{
			PrefetchValues: false,
		})
		defer it.Close()

		if limit == 0 {
			limit = maxResults // Maximum number of records to return
		}

		// Seek the prefix and check the key so we can quickly exit the iteration.
		for it.Seek(prefix); it.Valid() && message.ID(it.Item().Key()).EvalPrefix(ssid, cutoff) && len(matches) < maxResults && len(matches) < limit; it.Next() {
			//for it.Seek(prefix); it.Valid(); it.Next() {
			var msg message.Message
			item := it.Item()
			//k := item.Key()
			err := item.Value(func(v []byte) error {
				err = binary.Unmarshal(v, &msg)
				if err != nil {
					log.Error("adapter.Query", "unable to unmarshal value: "+err.Error())
					return err
				}
				matches = append(matches, msg)

				return nil
			})
			if err != nil {
				log.Error("adapter.Query", "unable to query db: "+err.Error())
				return err
			}
		}

		return nil
	}); err != nil {
		log.Error("adapter.Query", "db query error: "+err.Error())
		return nil, err
	}
	return matches, nil
}

func init() {
	store.RegisterAdapter(adapterName, &adapter{})
}
