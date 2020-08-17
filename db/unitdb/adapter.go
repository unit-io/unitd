package adapter

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/unit-io/bpool"
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
	Dir           string `json:"dir,omitempty"`
	Size          int64  `json:"mem_size"`
	LogReleaseDur string `json:"log_release_duration,omitempty"`
	dur           time.Duration
}

const (
	// Maximum number of records to return
	maxResults = 1024
	// Maximum TTL for message
	maxTTL = "24h"
)

// Store represents an SSD-optimized storage store.
type adapter struct {
	db         *unitdb.DB // The underlying database to store messages.
	mem        *memdb.DB  // The underlying memdb to store messages.
	config     *configType
	writeLockC chan struct{}
	bufPool    *bpool.BufferPool
	//tiny Batch
	tinyBatch *tinyBatch
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
	a.mem, err = memdb.Open(config.Size)
	if err != nil {
		return err
	}

	a.bufPool = bpool.NewBufferPool(config.Size, nil)
	a.tinyBatch.buffer = a.bufPool.Get()
	dur, err := time.ParseDuration(config.LogReleaseDur)
	if err != nil {
		return err
	}
	a.config = &config
	a.config.dur = dur
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
	entry := unitdb.NewEntry(topic, payload)
	entry.WithContract(contract)
	return a.db.PutEntry(entry)
}

// PutWithID appends the messages to the store using a pre generated messageId.
func (a *adapter) PutWithID(contract uint32, messageId, topic, payload []byte) error {
	entry := unitdb.NewEntry(topic, payload)
	entry.WithContract(contract)
	return a.db.PutEntry(entry.WithID(messageId))
}

// Get performs a query and attempts to fetch last n messages where
// n is specified by limit argument. From and until times can also be specified
// for time-series retrieval.
func (a *adapter) Get(contract uint32, topic []byte) (matches [][]byte, err error) {
	// Iterating over key/value pairs.
	query := unitdb.NewQuery(topic)
	query.WithContract(contract)
	return a.db.Get(query)
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
	entry := unitdb.NewEntry(topic, nil)
	entry.WithContract(contract)
	return a.db.DeleteEntry(entry.WithID(messageId))
}

type (
	tinyBatchInfo struct {
		entryCount uint32
	}

	tinyBatch struct {
		tinyBatchInfo
		buffer *bpool.Buffer
	}
)

func (b *tinyBatch) reset() {
	b.entryCount = 0
	atomic.StoreUint32(&b.entryCount, 0)
}

func (b *tinyBatch) count() uint32 {
	return atomic.LoadUint32(&b.entryCount)
}

func (b *tinyBatch) incount() uint32 {
	return atomic.AddUint32(&b.entryCount, 1)
}

// append appends message to tinyBatch for writing to log file.
func (a *adapter) Append(delFlag bool, k uint64, data []byte) error {
	var dBit uint8
	if delFlag {
		dBit = 1
	}
	var scratch [4]byte
	binary.LittleEndian.PutUint32(scratch[0:4], uint32(len(data)+8+4+1))

	if _, err := a.tinyBatch.buffer.Write(scratch[:]); err != nil {
		return err
	}

	// key with flag bit
	var key [9]byte
	key[0] = dBit
	binary.LittleEndian.PutUint64(key[1:], k)
	if _, err := a.tinyBatch.buffer.Write(key[:]); err != nil {
		return err
	}
	if data != nil {
		if _, err := a.tinyBatch.buffer.Write(data); err != nil {
			return err
		}
	}

	a.tinyBatch.incount()
	return nil
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

// Recovery recovers pending messages from log file.
func (a *adapter) Recovery(reset bool) (map[uint64][]byte, error) {
	m := make(map[uint64][]byte) // map[key]msg
	logOpts := wal.Options{Path: a.config.Dir + "/" + defaultMessageStore + logPostfix, TargetSize: a.config.Size, BufferSize: a.config.Size, Reset: reset}
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
	err = r.Read(func(last bool) (ok bool, err error) {
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

// Write writes tiny batch to log file
func (a *adapter) Write() error {
	if a.tinyBatch.count() == 0 {
		return nil
	}

	logWriter, err := a.wal.NewWriter()
	if err != nil {
		return err
	}
	// commit writes batches into write ahead log. The write happen synchronously.
	a.writeLockC <- struct{}{}
	defer func() {
		a.tinyBatch.buffer.Reset()
		<-a.writeLockC
	}()
	offset := uint32(0)
	buf := a.tinyBatch.buffer.Bytes()
	for i := uint32(0); i < a.tinyBatch.count(); i++ {
		dataLen := binary.LittleEndian.Uint32(buf[offset : offset+4])
		data := buf[offset+4 : offset+dataLen]
		if err := <-logWriter.Append(data); err != nil {
			return err
		}
		offset += dataLen
	}

	if err := <-logWriter.SignalInitWrite(nextTimeID(a.config.dur)); err != nil {
		return err
	}
	a.tinyBatch.reset()
	// signal log applied for older messages those are either acknowledged or timed out.
	return a.wal.SignalLogApplied(timeID(a.config.dur))
}

func timeID(dur time.Duration) int64 {
	return time.Now().UTC().Truncate(dur).Round(time.Millisecond).Unix()
}

func nextTimeID(dur time.Duration) int64 {
	return time.Now().UTC().Truncate(dur).Add(dur).Round(time.Millisecond).Unix()
}

func init() {
	adp := &adapter{
		writeLockC: make(chan struct{}),
		tinyBatch:  &tinyBatch{},
	}
	store.RegisterAdapter(adapterName, adp)
}
