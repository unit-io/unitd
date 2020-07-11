package adapter

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/unit-io/unitd/pkg/log"
	"github.com/unit-io/unitd/store"
	"github.com/unit-io/unitdb"
	"github.com/unit-io/unitdb/memdb"
	"github.com/unit-io/unitdb/wal"
)

const (
	defaultDatabase     = "unitd"
	defaultMessageStore = "messages"

	dbVersion = 2.0

	adapterName = "unitdb"

	logPostfix = ".log"
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
	db        *unitdb.DB // The underlying database to store messages.
	mem       *memdb.DB  // The underlying memdb to store messages.
	logWriter *wal.Writer
	wal       *wal.WAL
	version   int

	// close
	closer io.Closer
}

// Open initializes database connection
func (a *adapter) Open(jsonconfig string) error {
	if a.db != nil {
		return errors.New("unitdb adapter is already connected")
	}

	var err error
	var config configType

	if err = json.Unmarshal([]byte(jsonconfig), &config); err != nil {
		return errors.New("unitdb adapter failed to parse config: " + err.Error())
	}

	// Make sure we have a directory
	if err := os.MkdirAll(config.Dir, 0777); err != nil {
		log.Error("adapter.Open", "Unable to create db dir")
	}

	// Attempt to open the database
	a.db, err = unitdb.Open(config.Dir+"/"+defaultDatabase, nil, unitdb.WithMutable())
	if err != nil {
		log.Error("adapter.Open", "Unable to open db")
		return err
	}
	// Attempt to open the memdb
	a.mem, err = memdb.Open(1 << 33)
	if err != nil {
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
	if a.mem != nil {
		err = a.mem.Close()
		a.mem = nil

		var err error
		if a.closer != nil {
			if err1 := a.closer.Close(); err == nil {
				err = err1
			}
			a.closer = nil
		}
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
	err := a.db.PutEntry(unitdb.NewEntry(topic).WithPayload(payload).WithContract(contract))
	return err
}

// PutWithID appends the messages to the store using a pre generated messageId.
func (a *adapter) PutWithID(contract uint32, messageId, topic, payload []byte) error {
	err := a.db.PutEntry(unitdb.NewEntry(topic).WithID(messageId).WithPayload(payload).WithContract(contract))
	return err
}

// Get performs a query and attempts to fetch last n messages where
// n is specified by limit argument. From and until times can also be specified
// for time-series retrieval.
func (a *adapter) Get(contract uint32, topic []byte) (matches [][]byte, err error) {
	// Iterating over key/value pairs.
	matches, err = a.db.Get(unitdb.NewQuery(topic).WithContract(contract))
	if err != nil {
		return nil, err
	}
	return matches, nil
}

// NewID generates a new messageId.
func (a *adapter) NewID() ([]byte, error) {
	id := a.db.NewID()
	if id == nil {
		return nil, errors.New("Key is empty.")
	}
	return id, nil
}

// Put appends the messages to the store.
func (a *adapter) Delete(contract uint32, messageId, topic []byte) error {
	err := a.db.DeleteEntry(unitdb.NewEntry(topic).WithID(messageId).WithContract(contract))
	return err
}

// PutMessage appends the messages to the store.
func (a *adapter) PutMessage(blockId, key uint64, payload []byte) error {
	if err := a.mem.Set(blockId, key, payload); err != nil {
		return err
	}
	return nil
}

// GetMessage performs a query and attempts to fetch message for the given blockId and key
func (a *adapter) GetMessage(blockId, key uint64) (matches []byte, err error) {
	matches, err = a.mem.Get(blockId, key)
	if err != nil {
		return nil, err
	}
	return matches, nil
}

// Keys performs a query and attempts to fetch all keys for given blockId.
func (a *adapter) Keys(blockId uint64) []uint64 {
	return a.mem.Keys(blockId)
}

// DeleteMessage deletes message from memdb store.
func (a *adapter) DeleteMessage(blockId, key uint64) error {
	if err := a.mem.Remove(blockId, key); err != nil {
		return err
	}
	return nil
}

// NewWriter creates new log writer.
func (a *adapter) NewWriter() error {
	if w, err := a.wal.NewWriter(); err == nil {
		a.logWriter = w
		return err
	}
	return nil
}

// Append appends messages to the log.
func (a *adapter) Append(data []byte) <-chan error {
	return a.logWriter.Append(data)
}

// SignalInitWrite signals to write log.
func (a *adapter) SignalInitWrite(seq uint64) <-chan error {
	return a.logWriter.SignalInitWrite(seq)
}

// SignalLogApplied signals log has been applied for given upper sequence.
// logs are released from wal so that space can be reused.
func (a *adapter) SignalLogApplied(seq uint64) error {
	return a.wal.SignalLogApplied(seq)
}

// Recovery recovers pending messages from log file.
func (a *adapter) Recovery(path string, size int64, reset bool) (map[uint64][]byte, error) {
	m := make(map[uint64][]byte) // map[key]msg
	logOpts := wal.Options{Path: path + "/" + defaultMessageStore + logPostfix, TargetSize: size, BufferSize: size, Reset: reset}
	wal, needLogRecovery, err := wal.New(logOpts)
	if err != nil {
		wal.Close()
		return m, err
	}

	a.closer = wal
	a.wal = wal
	if !needLogRecovery || reset {
		return m, nil
	}

	// start log recovery
	r, err := wal.NewReader()
	if err != nil {
		return m, err
	}
	err = r.Read(func(lSeq uint64, last bool) (ok bool, err error) {
		l := r.Count()
		for i := uint32(0); i < l; i++ {
			logData, ok, err := r.Next()
			if err != nil {
				return false, err
			}
			if !ok {
				break
			}
			dBit := logData[0]
			key := binary.LittleEndian.Uint64(logData[1:9])
			msg := logData[9:]
			if dBit == 1 {
				if _, exists := m[key]; exists {
					delete(m, key)
				}
			}
			m[key] = msg
		}
		return false, nil
	})

	return m, err
}

func init() {
	store.RegisterAdapter(adapterName, &adapter{})
}
