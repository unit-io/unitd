package broker

import (
	"bufio"
	"context"
	"encoding/json"
	"time"

	lp "github.com/unit-io/unitd/lineprotocol"
	"github.com/unit-io/unitd/message"
	"github.com/unit-io/unitd/message/security"
	"github.com/unit-io/unitd/pkg/crypto"
	"github.com/unit-io/unitd/pkg/log"
	"github.com/unit-io/unitd/pkg/stats"
	"github.com/unit-io/unitd/pkg/uid"
	"github.com/unit-io/unitd/store"
	"github.com/unit-io/unitd/types"
)

const (
	requestClientId = 2682859131 // hash("clientid")
	requestKeygen   = 812942072  // hash("keygen")
)

func (c *Conn) readLoop() error {
	defer func() {
		log.Info("conn.Handler", "closing...")
		c.close()
	}()

	reader := bufio.NewReaderSize(c.socket, 65536)

	for {
		// Set read/write deadlines so we can close dangling connections
		c.socket.SetDeadline(time.Now().Add(time.Second * 120))

		// Decode an incoming packet
		pkt, err := lp.ReadPacket(c.proto, reader)
		if err != nil {
			return err
		}

		// Message handler
		if err := c.handler(pkt); err != nil {
			return err
		}
	}
}

// handle handles inbound packets.
func (c *Conn) handler(pkt lp.Packet) error {
	start := time.Now()
	var status int = 200
	defer func() {
		c.service.meter.ConnTimeSeries.AddTime(time.Since(start))
		c.service.stats.PrecisionTiming("conn_time_ns", time.Since(start), stats.IntTag("status", status))
	}()

	// Persist incoming
	c.storeInbound(pkt)

	switch pkt.Type() {
	// An attempt to connect.
	case lp.CONNECT:
		var returnCode uint8
		packet := *pkt.(*lp.Connect)

		c.insecure = packet.InsecureFlag
		c.username = string(packet.Username)
		clientid, err := c.onConnect(packet.ClientID)
		if err != nil {
			status = err.Status
			c.notifyError(err, 0)
			c.sendClientID(clientid.Encode(c.service.MAC))
			returnCode = 0x05 // Unauthorized
		}

		c.clientid = clientid
		c.MessageIds.Reset(message.MID(c.connid))
		// Take care of any messages in the store
		if !packet.CleanSessFlag {
			c.resume()
		} else {
			// contract is used as blockId and key prefix
			store.Log.Reset(c.clientid.Contract())
		}
		// Write the ack
		connack := &lp.Connack{ReturnCode: returnCode, ConnID: uint32(c.connid)}
		c.send <- connack

	// An attempt to subscribe to a topic.
	case lp.SUBSCRIBE:
		packet := *pkt.(*lp.Subscribe)
		ack := &lp.Suback{
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
		c.send <- ack

	// An attempt to unsubscribe from a topic.
	case lp.UNSUBSCRIBE:
		packet := *pkt.(*lp.Unsubscribe)
		ack := &lp.Unsuback{MessageID: packet.MessageID}

		// Unsubscribe from each subscription
		for _, sub := range packet.Topics {
			if err := c.onUnsubscribe(packet, sub.Topic); err != nil {
				status = err.Status
				c.notifyError(err, packet.MessageID)
			}
		}

		c.send <- ack

	// Ping response, respond appropriately.
	case lp.PINGREQ:
		resp := &lp.Pingresp{}
		c.send <- resp

	case lp.DISCONNECT:

	case lp.PUBLISH:
		packet := *pkt.(*lp.Publish)
		if err := c.onPublish(packet, packet.MessageID, packet.Topic, packet.Payload); err != nil {
			status = err.Status
			c.notifyError(err, packet.MessageID)
		}

	case lp.PUBREC:
		c.send <- pkt

	case lp.PUBREL:
		// persist outbound
		c.storeOutbound(pkt)
		c.send <- pkt

	case lp.PUBCOMP:
	}

	return nil
}

// writeLook handles outbound packets
func (c *Conn) writeLoop(ctx context.Context) {
	c.closeW.Add(1)
	defer c.closeW.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.closeC:
			return
		case msg, ok := <-c.pub:
			if !ok {
				// Channel closed.
				return
			}
			m, err := lp.Encode(c.proto, msg)
			if err != nil {
				log.Error("conn.writeLoop", err.Error())
				return
			}
			c.socket.Write(m.Bytes())
		case msg, ok := <-c.send:
			if !ok {
				// Channel closed.
				return
			}
			m, err := lp.Encode(c.proto, msg)
			if err != nil {
				log.Error("conn.writeLoop", err.Error())
				return
			}
			c.socket.Write(m.Bytes())
		}
	}
}

// onConnect is a handler for Connect events.
func (c *Conn) onConnect(clientID []byte) (uid.ID, *types.Error) {
	start := time.Now()
	defer log.ErrLogger.Debug().Str("context", "conn.onConnect").Int64("duration", time.Since(start).Nanoseconds()).Msg("")
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

// onSubscribe is a handler for Subscribe events.
func (c *Conn) onSubscribe(pkt lp.Subscribe, msgTopic []byte) *types.Error {
	start := time.Now()
	defer log.ErrLogger.Debug().Str("context", "conn.onSubscribe").Int64("duration", time.Since(start).Nanoseconds()).Msg("")

	//Parse Key
	topic := security.ParseKey(msgTopic)
	if topic.TopicType == security.TopicInvalid {
		return types.ErrBadRequest
	}

	if !c.insecure {
		if _, err := c.onSecureRequest(topic); err != nil {
			return err
		}
	}

	c.subscribe(pkt, topic)

	// if t0, t1, limit, ok := topic.Last(); ok {
	msgs, err := store.Message.Get(c.clientid.Contract(), topic.Topic)
	if err != nil {
		log.Error("conn.OnSubscribe", "query last messages"+err.Error())
		return types.ErrServerError
	}

	// Range over the messages in the channel and forward them
	for _, m := range msgs {
		msg := m // Copy message
		c.SendMessage(&msg)
	}

	return nil
}

// ------------------------------------------------------------------------------------

// onUnsubscribe is a handler for Unsubscribe events.
func (c *Conn) onUnsubscribe(pkt lp.Unsubscribe, msgTopic []byte) *types.Error {
	start := time.Now()
	defer log.ErrLogger.Debug().Str("context", "conn.onUnsubscribe").Int64("duration", time.Since(start).Nanoseconds()).Msg("")

	//Parse the key
	topic := security.ParseKey(msgTopic)
	if topic.TopicType == security.TopicInvalid {
		return types.ErrBadRequest
	}

	if !c.insecure {
		if _, err := c.onSecureRequest(topic); err != nil {
			return err
		}
	}

	c.unsubscribe(pkt, topic)

	return nil
}

// OnPublish is a handler for Publish events.
func (c *Conn) onPublish(pkt lp.Publish, messageID uint16, msgTopic []byte, payload []byte) *types.Error {
	start := time.Now()
	defer log.ErrLogger.Debug().Str("context", "conn.onPublish").Int64("duration", time.Since(start).Nanoseconds()).Msg("")

	//Parse the Key
	topic := security.ParseKey(msgTopic)
	if topic.TopicType == security.TopicInvalid {
		return types.ErrBadRequest
	}

	// Check whether the key is 'unitd' which means it's an API request
	if len(topic.Key) == 5 && string(topic.Key) == "unitd" {
		c.onUnitdRequest(topic, payload)
		return nil
	}

	if !c.insecure {
		wildcard, err := c.onSecureRequest(topic)
		if err != nil {
			return err
		}
		if wildcard {
			return types.ErrForbidden
		}
	}

	err := store.Message.Put(c.clientid.Contract(), topic.Topic, payload)
	if err != nil {
		log.Error("conn.onPublish", "store message "+err.Error())
		return types.ErrServerError
	}
	// Iterate through all subscribers and send them the message
	c.publish(pkt, messageID, topic, payload)

	// acknowledge a packet
	return c.ack(pkt)
}

// ack acknowledges a packet
func (c *Conn) ack(pkt lp.Publish) *types.Error {
	switch pkt.FixedHeader.Qos {
	case 2:
		pubrec := &lp.Pubrec{MessageID: pkt.MessageID}
		c.send <- pubrec
	case 1:
		puback := &lp.Puback{MessageID: pkt.MessageID}
		// persist outbound
		c.storeOutbound(puback)
		c.send <- puback
	case 0:
		// do nothing, since there is no need to send an ack packet back
	}
	return nil
}

// Load all stored messages and resend them to ensure QOS > 1,2 even after an application crash.
func (c *Conn) resume() {
	// contract is used as blockId and key prefix
	keys := store.Log.Keys(c.clientid.Contract())
	for _, k := range keys {
		msg := store.Log.Get(c.proto, k)
		if msg == nil {
			continue
		}
		// isKeyOutbound
		if (k & (1 << 4)) == 0 {
			switch msg.(type) {
			case *lp.Pubrel:
				c.send <- msg
			case *lp.Publish:
				c.send <- msg
			default:
				store.Log.Delete(k)
			}
		} else {
			switch msg.(type) {
			case *lp.Pubrel:
				c.recv <- msg
			default:
				store.Log.Delete(k)
			}
		}
	}
}

func (c *Conn) onSecureRequest(topic *security.Topic) (bool, *types.Error) {
	// Attempt to decode the key
	key, err := security.DecodeKey(topic.Key)
	if err != nil {
		return false, types.ErrBadRequest
	}

	// Check if the key has the permission to read the topic
	if !key.HasPermission(security.AllowRead) {
		return false, types.ErrUnauthorized
	}

	// Check if the key has the permission for the topic
	ok, wildcard := key.ValidateTopic(c.clientid.Contract(), topic.Topic[:topic.Size])
	if !ok {
		return wildcard, types.ErrUnauthorized
	}
	return wildcard, nil
}

// onSpecialRequest processes an special request.
func (c *Conn) onUnitdRequest(topic *security.Topic, payload []byte) (ok bool) {
	var resp interface{}
	defer func() {
		if b, err := json.Marshal(resp); err == nil {
			c.SendMessage(&message.Message{
				Topic:   []byte("unitd/" + string(topic.Topic[:topic.Size])),
				Payload: b,
			})
		}
	}()

	// Check query
	resp = types.ErrNotFound
	if len(topic.Topic[:topic.Size]) < 1 {
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
