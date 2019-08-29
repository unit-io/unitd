package broker

import (
	"bufio"
	"encoding/json"
	"errors"
	"time"

	"github.com/frontnet/trace/message"
	"github.com/frontnet/trace/message/security"
	"github.com/frontnet/trace/mqtt"
	"github.com/frontnet/trace/pkg/crypto"
	"github.com/frontnet/trace/pkg/log"
	"github.com/frontnet/trace/pkg/stats"
	"github.com/frontnet/trace/pkg/uid"
	"github.com/frontnet/trace/store"
	"github.com/frontnet/trace/types"
)

const (
	requestClientId = 2682859131 // hash("clientid")
	requestKeygen   = 812942072  // hash("keygen")
)

func (c *Conn) Handler() error {
	defer func() {
		log.Info("conn.Handler", "closing...")
		c.close()
	}()

	reader := bufio.NewReaderSize(c.socket, 65536)

	for {
		// Set read/write deadlines so we can close dangling connections
		c.socket.SetDeadline(time.Now().Add(time.Second * 120))

		// Decode an incoming MQTT packet
		pkt, err := mqtt.DecodePacket(reader)
		if err != nil {
			return err
		}

		// Mqtt message handler
		if err := c.handler(pkt); err != nil {
			return err
		}
	}
}

// handle handles an MQTT receive.
func (c *Conn) handler(pkt mqtt.Packet) error {
	start := time.Now()
	var status int = 200
	defer func() {
		c.service.meter.ConnTimeSeries.AddTime(time.Since(start))
		c.service.stats.PrecisionTiming("conn_time_ns", time.Since(start), stats.IntTag("status", status))
	}()

	switch pkt.Type() {
	// An attempt to connect to MQTT.
	case mqtt.CONNECT:
		packet := pkt.(*mqtt.Connect)
		c.username = string(packet.Username)
		clientid, err := c.onConnect(packet.ClientID)
		if err != nil {
			status = err.Status
			c.notifyError(err, 0)
			c.sendClientID(clientid.Encode(c.service.MAC))
		}

		c.clientid = clientid

		// Write the ack
		ack := mqtt.Connack{ReturnCode: 0x00}
		if !c.SendRawBytes(ack.Encode()) {
			return errors.New("conn.handler: The network connection timeout.")
		}

		return nil

		// An attempt to subscribe to a topic.
	case mqtt.SUBSCRIBE:
		packet := pkt.(*mqtt.Subscribe)
		ack := mqtt.Suback{
			MessageID: packet.MessageID,
			Qos:       make([]uint8, 0, len(packet.Subscriptions)),
		}

		// Subscribe for each subscription
		for _, sub := range packet.Subscriptions {
			if err := c.onSubscribe(packet, sub.Topic); err != nil {
				status = err.Status
				ack.Qos = append(ack.Qos, 0x80) // 0x80 indicate subscription failure
				c.notifyError(err, packet.MessageID)
				continue
			}

			// Append the QoS
			ack.Qos = append(ack.Qos, sub.Qos)
		}

		if packet.IsForwarded {
			return nil
		}

		// Acknowledge the subscription
		if !c.SendRawBytes(ack.Encode()) {
			return errors.New("conn.handler: The network connection timeout.")
		}

		return nil

	// An attempt to unsubscribe from a topic.
	case mqtt.UNSUBSCRIBE:
		packet := pkt.(*mqtt.Unsubscribe)
		ack := mqtt.Unsuback{MessageID: packet.MessageID}

		// Unsubscribe from each subscription
		for _, sub := range packet.Topics {
			if err := c.onUnsubscribe(packet, sub.Topic); err != nil {
				status = err.Status
				c.notifyError(err, packet.MessageID)
			}
		}

		// Acknowledge the unsubscription
		if !c.SendRawBytes(ack.Encode()) {
			return errors.New("conn.handler: The network connection timeout.")
		}

		return nil

	// MQTT ping response, respond appropriately.
	case mqtt.PINGREQ:
		ack := mqtt.Pingresp{}
		if !c.SendRawBytes(ack.Encode()) {
			return errors.New("conn.handler: The network connection timeout.")
		}
		return nil

	case mqtt.DISCONNECT:
		return nil

	case mqtt.PUBLISH:
		packet := pkt.(*mqtt.Publish)

		if err := c.onPublish(packet, packet.Topic, packet.Payload); err != nil {
			status = err.Status
			c.notifyError(err, packet.MessageID)
		}

		// Acknowledge the publication
		if packet.Header.QOS > 0 {
			ack := mqtt.Puback{MessageID: packet.MessageID}
			if !c.SendRawBytes(ack.Encode()) {
				return errors.New("conn.handler: The network connection timeout.")
			}
		}
		return nil
	}

	return nil
}

// onConnect is a handler for MQTT Connect events.
func (c *Conn) onConnect(clientID []byte) (uid.ID, *types.Error) {
	start := time.Now()
	defer log.ErrLogger.Debug().Str("context", "conn.onConnect").Dur("duration", time.Since(start)).Msg("")
	var clientid = uid.ID{}
	if clientID != nil && len(clientID) > c.service.MAC.Overhead() {
		if contract, ok := c.service.cache.Load(crypto.SignatureToUint32(clientID[crypto.EpochSize:crypto.MessageOffset])); ok {
			clientid, err := uid.CachedClientID(contract.(uint32))
			if err != nil {
				return nil, types.ErrUnauthorized
			}
			return clientid, nil
		}
	}

	clientid, err := uid.Decode(clientID, c.service.MAC)

	if err != nil {
		clientid, err = uid.NewClientID(1)
		if err != nil {
			return nil, types.ErrUnauthorized
		}

		return clientid, types.ErrInvalidClientId
	}

	//do not cache primary client Id
	if !clientid.IsPrimary() {
		cid := []byte(clientid.Encode(c.service.MAC))
		c.service.cache.LoadOrStore(crypto.SignatureToUint32(cid[crypto.EpochSize:crypto.MessageOffset]), clientid.Contract())
	}

	return clientid, nil
}

// onSubscribe is a handler for MQTT Subscribe events.
func (c *Conn) onSubscribe(pkt *mqtt.Subscribe, mqttTopic []byte) *types.Error {
	start := time.Now()
	defer log.ErrLogger.Debug().Str("context", "conn.onSubscribe").Dur("duration", time.Since(start)).Msg("")

	//Parse Key
	topic := message.ParseKey(mqttTopic)

	// Attempt to decode the key
	key, err := security.DecodeKey(topic.Key)
	if err != nil {
		return types.ErrBadRequest
	}

	// Check if the key has the permission to read the topic
	if !key.HasPermission(security.AllowRead) {
		return types.ErrUnauthorized
	}

	// Check if the key has the permission for the topic
	ok, wildcard := key.ValidateTopic(c.clientid.Contract(), topic.Topic)
	if !ok {
		return types.ErrUnauthorized
	}

	// Parse the topic
	topic.Parse(c.clientid.Contract(), wildcard)
	if topic.TopicType == message.TopicInvalid {
		return types.ErrBadRequest
	}

	// Add contract to the parts
	message.AddContract(c.clientid.Contract(), topic)
	c.subscribe(pkt, topic)

	// In case of ttl, store messages to database
	if t0, t1, limit, ok := topic.Last(); ok {
		ssid := message.NewSsid(topic.Parts)
		msgs, err := store.Message.Query(ssid, t0, t1, int(limit))
		if err != nil {
			log.Error("conn.OnSubscribe", "query last messages"+err.Error())
			return types.ErrServerError
		}

		// Range over the messages in the channel and forward them
		for _, m := range msgs {
			msg := m // Copy message
			c.SendMessage(&msg)
		}
	}

	return nil
}

// ------------------------------------------------------------------------------------

// onUnsubscribe is a handler for MQTT Unsubscribe events.
func (c *Conn) onUnsubscribe(pkt *mqtt.Unsubscribe, mqttTopic []byte) *types.Error {
	start := time.Now()
	defer log.ErrLogger.Debug().Str("context", "conn.onUnsubscribe").Dur("duration", time.Since(start)).Msg("")

	//Parse the key
	topic := message.ParseKey(mqttTopic)

	// Attempt to decode the key
	key, err := security.DecodeKey(topic.Key)
	if err != nil {
		return types.ErrBadRequest
	}

	// Check if the key has the permission to read the topic
	if !key.HasPermission(security.AllowRead) {
		return types.ErrUnauthorized
	}

	// Check if the key has the permission for the topic
	ok, wildcard := key.ValidateTopic(c.clientid.Contract(), topic.Topic)
	if !ok {
		return types.ErrUnauthorized
	}

	// Parse the topic
	topic.Parse(c.clientid.Contract(), wildcard)
	if topic.TopicType == message.TopicInvalid {
		return types.ErrBadRequest
	}

	// Add contract to the parts
	message.AddContract(c.clientid.Contract(), topic)
	c.unsubscribe(pkt, topic)

	return nil
}

// OnPublish is a handler for MQTT Publish events.
func (c *Conn) onPublish(pkt *mqtt.Publish, mqttTopic []byte, payload []byte) *types.Error {
	start := time.Now()
	defer log.ErrLogger.Debug().Str("context", "conn.onPublish").Dur("duration", time.Since(start)).Msg("")

	//Parse the Key
	topic := message.ParseKey(mqttTopic)
	// Check whether the key is 'trace' which means it's an API request
	if len(topic.Key) == 5 && string(topic.Key) == "trace" {
		topic.Parse(message.Contract, false)
		c.onTraceRequest(topic, payload)
		return nil
	}

	//Try to decode the key
	key, err := security.DecodeKey(topic.Key)
	if err != nil {
		return types.ErrBadRequest
	}

	// Check if the key has the permission to read the topic
	if !key.HasPermission(security.AllowWrite) {
		return types.ErrUnauthorized
	}

	// Check if the key has the permission for the topic
	ok, wildcard := key.ValidateTopic(c.clientid.Contract(), topic.Topic)
	if !ok {
		return types.ErrUnauthorized
	}

	if wildcard {
		return types.ErrForbidden
	}

	// Parse the topic
	topic.Parse(c.clientid.Contract(), false)
	if topic.TopicType == message.TopicInvalid {
		return types.ErrBadRequest
	}

	// Publish should only have static topic strings
	if topic.TopicType != message.TopicStatic {
		return types.ErrForbidden
	}

	// Add contract to the parts
	message.AddContract(c.clientid.Contract(), topic)
	ssid := message.NewSsid(topic.Parts)

	// Create a new message
	msg := message.New(
		ssid,
		topic.Topic,
		payload,
	)

	// In case of ttl, add ttl to the msg and store to the db
	if ttl, ok := topic.TTL(); ok {
		//1410065408 10 sec
		msg.TTL = ttl // Add the TTL to the message
		store.Message.Store(msg)
	}

	// Iterate through all subscribers and send them the message
	c.publish(pkt, topic, msg)

	return nil
}

// onSpecialRequest processes an special request.
func (c *Conn) onTraceRequest(topic *message.Topic, payload []byte) (ok bool) {
	var resp interface{}
	defer func() {
		if b, err := json.Marshal(resp); err == nil {
			c.SendMessage(&message.Message{
				Topic:   []byte("trace/" + string(topic.Topic)),
				Payload: b,
			})
		}
	}()

	// Check query
	resp = types.ErrNotFound
	if len(topic.Parts) < 1 {
		return
	}

	switch topic.Target() {
	case requestClientId:
		resp, ok = c.onClientIdRequest()
		return
	case requestKeygen:
		resp, ok = c.onKeyGen(payload)
		return
	default:
		return
	}
}

// onClientIdRequest is a handler that returns new client id for the request.
func (c *Conn) onClientIdRequest() (interface{}, bool) {
	if !c.clientid.IsPrimary() {
		return types.ErrClientIdForbidden, false
	}

	clientid, err := uid.NewSecondaryClientID(c.clientid)
	if err != nil {
		return types.ErrBadRequest, false
	}
	cid := clientid.Encode(c.service.MAC)
	return &types.ClientIdResponse{
		Status:   200,
		ClientId: cid,
	}, true

}

// onKeyGen processes a keygen request.
func (c *Conn) onKeyGen(payload []byte) (interface{}, bool) {
	// Deserialize the payload.
	msg := types.KeyGenRequest{}
	if err := json.Unmarshal(payload, &msg); err != nil {
		return types.ErrBadRequest, false
	}

	// Use the cipher to generate the key
	key, err := security.GenerateKey(c.clientid.Contract(), []byte(msg.Topic), msg.Access())
	if err != nil {
		switch err {
		case security.ErrTargetTooLong:
			return types.ErrTargetTooLong, false
		default:
			return types.ErrServerError, false
		}
	}

	// Success, return the response
	return &types.KeyGenResponse{
		Status: 200,
		Key:    key,
		Topic:  msg.Topic,
	}, true
}
