package broker

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	jcr "github.com/DisposaBoy/JsonConfigReader"
	"github.com/stretchr/testify/assert"
	"github.com/unit-io/unitd/config"
	"github.com/unit-io/unitd/mqtt"
)

func TestPubsub(t *testing.T) {
	var cfg *config.Config
	// Get the directory of the process
	// exe, err := os.Executable()
	_, exe, _, _ := runtime.Caller(0)
	configfile := filepath.Join(filepath.Dir(exe), "../unitd.conf")
	if file, err := os.Open(configfile); err != nil {
		assert.NoError(t, err)
	} else if err = json.NewDecoder(jcr.New(file)).Decode(&cfg); err != nil {
		assert.NoError(t, err)
	}
	svc, err := NewService(context.Background(), cfg)
	assert.NoError(t, err)

	defer svc.Close()

	go svc.Listen()

	// Create a client
	cli, err := net.Dial("tcp", "127.0.0.1:6060")
	assert.NoError(t, err)
	defer cli.Close()

	{ // Connect to the broker
		connect := mqtt.Connect{ClientID: []byte("UCBFDONCNJLaKMCAIeJBaOVfbAXUZHNPLDKKLDKLHZHKYIZLCDPQ")}
		n := connect.Encode()
		assert.Equal(t, 14, n)
		assert.NoError(t, err)
	}

	{ // Read connack
		pkt, err := mqtt.DecodePacket(cli)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.CONNACK, pkt.Type())
	}

	{ // Ping the broker
		ping := mqtt.Pingreq{}
		n := ping.Encode()
		assert.Equal(t, 2, n)
		assert.NoError(t, err)
	}

	{ // Read pong
		pkt, err := mqtt.DecodePacket(cli)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.PINGRESP, pkt.Type())
	}

	{ // Subscribe to a topic
		sub := mqtt.Subscribe{
			Header: &mqtt.FixedHeader{QOS: 0},
			Subscriptions: []mqtt.TopicQOSTuple{
				{Topic: []byte("AYAAMACRZDCHK/..."), Qos: 0},
			},
		}
		sub.Encode()
		assert.NoError(t, err)
	}

	{ // Read suback
		pkt, err := mqtt.DecodePacket(cli)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.SUBACK, pkt.Type())
	}

	{ // Publish a message
		msg := mqtt.Publish{
			Header:  &mqtt.FixedHeader{QOS: 0},
			Topic:   []byte("AbYANcEEZDcdY/unit8.b.b1?ttl=3m"),
			Payload: []byte("Hi unit8.b.b1!"),
		}
		msg.Encode()
		assert.NoError(t, err)
	}

	{ // Read the message back
		pkt, err := mqtt.DecodePacket(cli)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.PUBLISH, pkt.Type())
		assert.Equal(t, &mqtt.Publish{
			Header:  &mqtt.FixedHeader{QOS: 0},
			Topic:   []byte("unit8.b.b1"),
			Payload: []byte("Hi unit8.b.b1!"),
		}, pkt)
	}

	{ // Unsubscribe from the topic
		sub := mqtt.Unsubscribe{
			Header: &mqtt.FixedHeader{QOS: 0},
			Topics: []mqtt.TopicQOSTuple{
				{Topic: []byte("AYAAMACRZDCHK/..."), Qos: 0},
			},
		}
		sub.Encode()
		assert.NoError(t, err)
	}

	{ // Read unsuback
		pkt, err := mqtt.DecodePacket(cli)
		assert.NoError(t, err)
		assert.Equal(t, mqtt.UNSUBACK, pkt.Type())
	}

	{ // Disconnect from the broker
		disconnect := mqtt.Disconnect{}
		n := disconnect.Encode()
		assert.Equal(t, 2, n)
		assert.NoError(t, err)
	}

}
