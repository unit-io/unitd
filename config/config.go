package config

import (
	"encoding/json"

	"github.com/saffat-in/trace/pkg/log"
)

const (
	MaxMessageSize = 65536 // Maximum message size allowed from/to the peer.
)

// Config represents main configuration.
type Config struct {
	// Default HTTP(S) address:port to listen on for websocket and long polling clients. Either a
	// numeric or a canonical name, e.g. ":80" or ":https". Could include a host name, e.g.
	// "localhost:80".
	// Could be blank: if TLS is not configured, will use ":80", otherwise ":443".
	// Can be overridden from the command line, see option --listen.
	Listen string `json:"listen"`

	// Default logging level is "InfoLevel" so to enable the debug log set the "LogLevel" to "DebugLevel".
	LoggingLevel string `json:"logging_level"`

	// MaxMessageSize     int             `json:"max_message_size"`
	// // Maximum number of topic subscribers.
	// MaxSubscriberCount int             `json:"max_subscriber_count"`

	EncryptionConfig json.RawMessage `json:"encryption_config"`

	// Configs for subsystems
	Cluster json.RawMessage `json:"cluster_config"`

	// Config for database store
	Store json.RawMessage `json:"store_config"`

	// Config to expose runtime stats
	VarzPath string `json:"varz_path"`
}

// EncryptionConfig represents the configuration for the encryption.
type EncryptionConfig struct {

	// chacha20poly1305 encryption key for client Ids and topic keys. 32 random bytes base64-encoded.
	Key string `json:"key,omitempty"`

	// Key identifier. it is useful when you use multiple keys.
	Identifier string `json:"identifier"`

	// slealed flag tells if key in the configuration is sealed.
	Sealed bool `json:"slealed"`

	// timestamp is helpful to determine the latest key in case of keyroll over.
	Timestamp uint32 `json:"timestamp,omitempty"`
}

func (c *Config) Encryption(encrConfig json.RawMessage) EncryptionConfig {
	var encr EncryptionConfig
	if err := json.Unmarshal(encrConfig, &encr); err != nil {
		log.Fatal("config.Encryption", "error in parsing encryption config", err)
	}

	return encr
}
