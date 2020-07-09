package store

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/unit-io/bpool"
	adapter "github.com/unit-io/unitd/db"
	"github.com/unit-io/unitd/message"
	lp "github.com/unit-io/unitd/net/lineprotocol"
	"github.com/unit-io/unitd/pkg/log"
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

// SubscriptionStore is a Subscription struct to hold methods for persistence mapping for the subscription.
// Note, do not use same contract as messagestore
type SubscriptionStore struct{}

// Message is the ancor for storing/retrieving Message objects
var Subscription SubscriptionStore

func (s *SubscriptionStore) Put(contract uint32, messageId, topic, payload []byte) error {
	return adp.PutWithID(contract^connStoreId, messageId, topic, payload)
}

func (s *SubscriptionStore) Get(contract uint32, topic []byte) (matches [][]byte, err error) {
	resp, err := adp.Get(contract^connStoreId, topic)
	for _, payload := range resp {
		if payload == nil {
			continue
		}
		matches = append(matches, payload)
	}

	return matches, err
}

func (s *SubscriptionStore) NewID() ([]byte, error) {
	return adp.NewID()
}

func (s *SubscriptionStore) Delete(contract uint32, messageId, topic []byte) error {
	return adp.Delete(contract^connStoreId, messageId, topic)
}

// MessageStore is a Message struct to hold methods for persistence mapping for the Message object.
type MessageStore struct{}

// Message is the anchor for storing/retrieving Message objects
var Message MessageStore

func (m *MessageStore) Put(contract uint32, topic, payload []byte) error {
	return adp.Put(contract, topic, payload)
}

func (m *MessageStore) Get(contract uint32, topic []byte) (matches []message.Message, err error) {
	resp, err := adp.Get(contract, topic)
	for _, payload := range resp {
		msg := message.Message{
			Topic:   topic,
			Payload: payload,
		}
		matches = append(matches, msg)
	}

	return matches, err
}

type storeConfigType struct {
	Dir           string `json:"dir,omitempty"`
	Size          int64  `json:"mem_size"`
	LogReleaseDur string `json:"log_release_duration"`
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

// MessageLog is a Message struct to hold methods for persistence mapping for the Message object.
type MessageLog struct {
	seq        uint64
	writeLockC chan struct{}
	bufPool    *bpool.BufferPool
	//tiny Batch
	tinyBatch *tinyBatch
}

// Log is the anchor for storing/retrieving Message objects
var Log MessageLog

func recovery(path string, size int64, reset bool) error {
	m, err := adp.Recovery(path, size, reset)
	if err != nil {
		return err
	}
	for k, msg := range m {
		blockId := k & 0xFFFFFFFF
		if err := adp.PutMessage(blockId, k, msg); err != nil {
			return err
		}
	}
	return nil
}

// InitMessageStore init message store and start recovery if reset flag is not set.
func InitMessageStore(jsonconf string, reset bool) error {
	var config configType
	if err := json.Unmarshal([]byte(jsonconf), &config); err != nil {
		return errors.New("store: failed to parse config: " + err.Error() + "(" + jsonconf + ")")
	}

	var storeConfig storeConfigType
	if err := json.Unmarshal(config.Adapters[adp.GetName()], &storeConfig); err != nil {
		return errors.New("unitdb adapter failed to parse config: " + err.Error())
	}

	Log = MessageLog{
		writeLockC: make(chan struct{}, 1),
		bufPool:    bpool.NewBufferPool(storeConfig.Size, nil),
		tinyBatch:  &tinyBatch{},
	}
	Log.tinyBatch.buffer = Log.bufPool.Get()
	if err := recovery(storeConfig.Dir, storeConfig.Size, reset); err != nil {
		return err
	}
	Log.tinyBatchLoop(15 * time.Millisecond)
	dur, err := time.ParseDuration(storeConfig.LogReleaseDur)
	if err != nil {
		return err
	}
	logReleaser(dur)
	return nil
}

// PersistOutbound handles which outgoing messages are stored
func (l *MessageLog) PersistOutbound(proto lp.ProtoAdapter, key uint64, msg lp.Packet) {
	blockId := key & 0xFFFFFFFF
	switch msg.Info().Qos {
	case 0:
		switch msg.(type) {
		case *lp.Puback, *lp.Pubcomp:
			// Sending puback. delete matching publish
			// from ibound
			adp.DeleteMessage(blockId, key)
			l.append(true, key, nil)
		}
	case 1:
		switch msg.(type) {
		case *lp.Publish, *lp.Pubrel, *lp.Subscribe, *lp.Unsubscribe:
			// Sending publish. store in obound
			// until puback received
			m, err := lp.Encode(proto, msg)
			if err != nil {
				log.ErrLogger.Err(err).Str("context", "store.PersistOutbound")
				return
			}
			adp.PutMessage(blockId, key, m.Bytes())
			l.append(false, key, m.Bytes())
		default:
		}
	case 2:
		switch msg.(type) {
		case *lp.Publish:
			// Sending publish. store in obound
			// until pubrel received
			m, err := lp.Encode(proto, msg)
			if err != nil {
				log.ErrLogger.Err(err).Str("context", "store.PersistOutbound")
				return
			}
			adp.PutMessage(blockId, key, m.Bytes())
			l.append(false, key, m.Bytes())
		default:
		}
	}
}

// PersistInbound handles which incoming messages are stored
func (l *MessageLog) PersistInbound(proto lp.ProtoAdapter, key uint64, msg lp.Packet) {
	blockId := key & 0xFFFFFFFF
	switch msg.Info().Qos {
	case 0:
		switch msg.(type) {
		case *lp.Puback, *lp.Suback, *lp.Unsuback, *lp.Pubcomp:
			// Received a puback. delete matching publish
			// from obound
			adp.DeleteMessage(blockId, key)
			l.append(true, key, nil)
		case *lp.Publish, *lp.Pubrec, *lp.Connack:
		default:
		}
	case 1:
		switch msg.(type) {
		case *lp.Publish, *lp.Pubrel:
			// Received a publish. store it in ibound
			// until puback sent
			m, err := lp.Encode(proto, msg)
			if err != nil {
				log.ErrLogger.Err(err).Str("context", "store.PersistOutbound")
				return
			}
			adp.PutMessage(blockId, key, m.Bytes())
			l.append(false, key, m.Bytes())
		default:
		}
	case 2:
		switch msg.(type) {
		case *lp.Publish:
			// Received a publish. store it in ibound
			// until pubrel received
			m, err := lp.Encode(proto, msg)
			if err != nil {
				log.ErrLogger.Err(err).Str("context", "store.PersistOutbound")
				return
			}
			adp.PutMessage(blockId, key, m.Bytes())
			l.append(false, key, m.Bytes())
		default:
		}
	}
}

// Get performs a query and attempts to fetch message for the given blockId and key
func (l *MessageLog) Get(proto lp.ProtoAdapter, key uint64) lp.Packet {
	blockId := key & 0xFFFFFFFF
	if raw, err := adp.GetMessage(blockId, key); raw != nil && err == nil {
		r := bytes.NewReader(raw)
		if msg, err := lp.ReadPacket(proto, r); err == nil {
			return msg
		}

	}
	return nil
}

// Keys performs a query and attempts to fetch all keys for given blockId and key prefix.
func (l *MessageLog) Keys(prefix uint32) []uint64 {
	matches := make([]uint64, 0)
	keys := adp.Keys(uint64(prefix))
	for _, k := range keys {
		if evalPrefix(k, prefix) {
			matches = append(matches, k)
		}
	}
	return matches
}

// Delete is used to delete message.
func (l *MessageLog) Delete(key uint64) {
	blockId := key & 0xFFFFFFFF
	adp.DeleteMessage(blockId, key)
	l.append(true, key, nil)
}

// Reset removes all keys from store for the given blockId and key prefix
func (l *MessageLog) Reset(prefix uint32) {
	keys := adp.Keys(uint64(prefix))
	for _, k := range keys {
		if evalPrefix(k, prefix) {
			adp.DeleteMessage(uint64(prefix), k)
		}
	}
}

// append appends message to tinyBatch for writing to log file.
func (l *MessageLog) append(delFlag bool, k uint64, data []byte) error {
	var dBit uint8
	if delFlag {
		dBit = 1
	}
	var scratch [4]byte
	binary.LittleEndian.PutUint32(scratch[0:4], uint32(len(data)+8+4+1))

	if _, err := l.tinyBatch.buffer.Write(scratch[:]); err != nil {
		return err
	}

	// key with flag bit
	var key [9]byte
	key[0] = dBit
	binary.LittleEndian.PutUint64(key[1:], k)
	if _, err := l.tinyBatch.buffer.Write(key[:]); err != nil {
		return err
	}
	if data != nil {
		if _, err := l.tinyBatch.buffer.Write(data); err != nil {
			return err
		}
	}

	l.tinyBatch.incount()
	return nil
}

// tinyCommit commits tiny batch to log file
func (l *MessageLog) tinyCommit() error {
	if l.tinyBatch.count() == 0 {
		return nil
	}

	if err := adp.NewWriter(); err != nil {
		return err
	}
	// commit writes batches into write ahead log. The write happen synchronously.
	l.writeLockC <- struct{}{}
	defer func() {
		l.tinyBatch.buffer.Reset()
		<-l.writeLockC
	}()
	offset := uint32(0)
	buf := l.tinyBatch.buffer.Bytes()
	for i := uint32(0); i < l.tinyBatch.count(); i++ {
		dataLen := binary.LittleEndian.Uint32(buf[offset : offset+4])
		data := buf[offset+4 : offset+dataLen]
		if err := <-adp.Append(data); err != nil {
			return err
		}
		offset += dataLen
	}

	if err := <-adp.SignalInitWrite(timeNow()); err != nil {
		return err
	}
	l.tinyBatch.reset()
	return nil
}

func evalPrefix(key uint64, prefix uint32) bool {
	return uint64(prefix) == key&0xFFFFFFFF
}

func timeNow() uint64 {
	return uint64(time.Now().UTC().Round(time.Millisecond).Unix())
}

func timeSeq(dur time.Duration) uint64 {
	return uint64(time.Now().UTC().Truncate(dur).Round(time.Millisecond).Unix())
}

// tinyBatchLoop handles tiny bacthes write to log.
func (m *MessageLog) tinyBatchLoop(interval time.Duration) {
	ctx := context.Background()
	go func() {
		tinyBatchWriterTicker := time.NewTicker(interval)
		defer func() {
			tinyBatchWriterTicker.Stop()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tinyBatchWriterTicker.C:
				if err := m.tinyCommit(); err != nil {
					fmt.Println("Error committing tinyBatch")
				}
			}
		}
	}()
}

// logs are released from wal if older than a minute.
func logReleaser(dur time.Duration) {
	ctx := context.Background()
	go func() {
		logTicker := time.NewTicker(dur)
		defer logTicker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-logTicker.C:
				adp.SignalLogApplied(timeSeq(dur))
			}
		}
	}()
}
