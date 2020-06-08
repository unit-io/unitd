package broker

import (
	"encoding/json"
	"net"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/unit-io/unitd/message"
	"github.com/unit-io/unitd/message/security"
	"github.com/unit-io/unitd/pkg/log"
	"github.com/unit-io/unitd/pkg/uid"
	"github.com/unit-io/unitd/plugins/mqtt"
	"github.com/unit-io/unitd/store"
	"github.com/unit-io/unitd/types"
)

type Conn struct {
	sync.Mutex
	tracked uint32 // Whether the connection was already tracked or not.
	// protocol - NONE (unset), RPC, GRPC, WEBSOCK, CLUSTER
	proto    types.Proto
	socket   net.Conn
	send     chan []byte
	stop     chan interface{}
	username string         // The username provided by the client during MQTT connect.
	clientid uid.ID         // The clientid provided by client during MQTT connect or new Id assigned.
	connid   uid.LID        // The locally unique id of the connection.
	service  *Service       // The service for this connection.
	subs     *message.Stats // The subscriptions for this connection.
	// Reference to the cluster node where the connection has originated. Set only for cluster RPC sessions
	clnode *ClusterNode
	// Cluster nodes to inform when disconnected
	nodes map[string]bool
}

func (s *Service) newConn(t net.Conn, proto types.Proto) *Conn {
	c := &Conn{
		proto:   proto,
		socket:  t,
		send:    make(chan []byte, 1),      // buffered
		stop:    make(chan interface{}, 1), // Buffered by 1 just to make it non-blocking
		connid:  uid.NewLID(),
		service: s,
		subs:    message.NewStats(),
	}

	// Increment the connection counter
	s.meter.Connections.Inc(1)

	Globals.ConnCache.Add(c)
	return c
}

// newRpcConn a new connection in cluster
func (s *Service) newRpcConn(conn interface{}, connid uid.LID, clientid uid.ID) *Conn {
	c := &Conn{
		connid:   connid,
		clientid: clientid,
		send:     make(chan []byte),         // buffered
		stop:     make(chan interface{}, 1), // Buffered by 1 just to make it non-blocking
		service:  s,
		subs:     message.NewStats(),
		clnode:   conn.(*ClusterNode),
		nodes:    make(map[string]bool, 3),
	}

	Globals.ConnCache.Add(c)
	return c
}

// ID returns the unique identifier of the subsriber.
func (c *Conn) ID() string {
	return strconv.FormatUint(uint64(c.connid), 10)
}

// Type returns the type of the subscriber
func (c *Conn) Type() message.SubscriberType {
	return message.SubscriberDirect
}

// Send forwards the message to the underlying client.
func (c *Conn) SendMessage(m *message.Message) bool {
	packet := mqtt.Publish{
		Header: &mqtt.FixedHeader{
			QOS: 0, // TODO when we'll support more QoS
		},
		MessageID: 0,         // TODO
		Topic:     m.Topic,   // The topic for this message.
		Payload:   m.Payload, // The payload for this message.
	}

	// Acknowledge the publication
	select {
	case c.send <- packet.Encode():
	case <-time.After(time.Microsecond * 50):
		return false
	}

	return true
}

// Send forwards raw bytes to the underlying client.
func (c *Conn) SendRawBytes(buf []byte) bool {
	if c == nil {
		return true
	}

	select {
	case c.send <- buf:
	case <-time.After(time.Microsecond * 50):
		return false
	}

	return true
}

func (c *Conn) writeLoop() {
	defer func() {
		// Break readLoop.
		c.close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				// Channel closed.
				return
			}
			c.socket.Write(msg)
		}
	}
}

// Subscribe subscribes to a particular topic.
func (c *Conn) subscribe(pkt *mqtt.Subscribe, topic *security.Topic) (err error) {
	c.Lock()
	defer c.Unlock()

	key := string(topic.Key)
	if exists := c.subs.Exist(key); exists && !pkt.IsForwarded && Globals.Cluster.isRemoteContract(string(c.clientid.Contract())) {
		// The contract is handled by a remote node. Forward message to it.
		if err := Globals.Cluster.routeToContract(pkt, topic, message.SUBSCRIBE, &message.Message{}, c); err != nil {
			log.ErrLogger.Err(err).Str("context", "conn.subscribe").Int64("connid", int64(c.connid)).Msg("unable to subscribe to remote topic")
			return err
		}
		// Add the subscription to Counters
	} else {
		messageId, err := store.Connection.NewID()
		if err != nil {
			log.ErrLogger.Err(err).Str("context", "conn.subscribe")
		}
		if first := c.subs.Increment(topic.Topic[:topic.Size], key, messageId); first {
			// Subscribe the subscriber
			if err = store.Connection.Put(c.clientid.Contract(), topic.Topic, messageId, c.connid); err != nil {
				log.ErrLogger.Err(err).Str("context", "conn.subscribe").Str("topic", string(topic.Topic[:topic.Size])).Int64("connid", int64(c.connid)).Msg("unable to subscribe to topic") // Unable to subscribe
				return err
			}
			// Increment the subscription counter
			c.service.meter.Subscriptions.Inc(1)
		}
	}
	return nil
}

// Unsubscribe unsubscribes this client from a particular topic.
func (c *Conn) unsubscribe(pkt *mqtt.Unsubscribe, topic *security.Topic) (err error) {
	c.Lock()
	defer c.Unlock()

	key := string(topic.Key)
	// Remove the subscription from stats and if there's no more subscriptions, notify everyone.
	if last, messageId := c.subs.Decrement(topic.Topic[:topic.Size], key); last {
		// Unsubscribe the subscriber
		if err = store.Connection.Delete(c.clientid.Contract(), topic.Topic[:topic.Size], messageId); err != nil {
			log.ErrLogger.Err(err).Str("context", "conn.unsubscribe").Str("topic", string(topic.Topic[:topic.Size])).Int64("connid", int64(c.connid)).Msg("unable to unsubscribe to topic") // Unable to subscribe
			return err
		}
		// Decrement the subscription counter
		c.service.meter.Subscriptions.Dec(1)
	}
	if !pkt.IsForwarded && Globals.Cluster.isRemoteContract(string(c.clientid.Contract())) {
		// The topic is handled by a remote node. Forward message to it.
		if err := Globals.Cluster.routeToContract(pkt, topic, message.UNSUBSCRIBE, &message.Message{}, c); err != nil {
			log.ErrLogger.Err(err).Str("context", "conn.unsubscribe").Int64("connid", int64(c.connid)).Msg("unable to unsubscribe to remote topic")
			return err
		}
	}
	return nil
}

// Publish publishes a message to everyone and returns the number of outgoing bytes written.
func (c *Conn) publish(pkt *mqtt.Publish, topic *security.Topic, payload []byte) (err error) {
	c.service.meter.InMsgs.Inc(1)
	c.service.meter.InBytes.Inc(int64(len(payload)))
	// subsciption count
	scount := 0

	conns, err := store.Connection.Get(c.clientid.Contract(), topic.Topic)
	if err != nil {
		log.ErrLogger.Err(err).Str("context", "conn.publish")
	}
	m := &message.Message{
		Topic:   topic.Topic[:topic.Size],
		Payload: payload,
	}
	for _, connid := range conns {
		sub := Globals.ConnCache.Get(connid)
		if sub != nil {
			if !sub.SendMessage(m) {
				log.Error("conn.publish", "publish timeout")
			}
			scount++
		}
	}
	c.service.meter.OutMsgs.Inc(int64(scount))
	c.service.meter.OutBytes.Inc(m.Size() * int64(scount))

	if !pkt.IsForwarded && Globals.Cluster.isRemoteContract(string(c.clientid.Contract())) {
		if err = Globals.Cluster.routeToContract(pkt, topic, message.PUBLISH, m, c); err != nil {
			log.ErrLogger.Err(err).Str("context", "conn.publish").Int64("connid", int64(c.connid)).Msg("unable to publish to remote topic")
		}
	}
	return err
}

// sendClientID generate unique client and send it to new client
func (c *Conn) sendClientID(clientidentifier string) {
	c.SendMessage(&message.Message{
		Topic:   []byte("$SYS/client_identifier/"),
		Payload: []byte(clientidentifier),
	})
}

// notifyError notifies the connection about an error
func (c *Conn) notifyError(err *types.Error, messageID uint16) {
	err.ID = int(messageID)
	if b, err := json.Marshal(err); err == nil {
		c.SendMessage(&message.Message{
			Topic:   []byte("unitd/error/"),
			Payload: b,
		})
	}
}

func (c *Conn) unsubAll() {
	for _, stat := range c.subs.All() {
		store.Connection.Delete(c.clientid.Contract(), stat.Topic, stat.ID)
	}
}

// Close terminates the connection.
func (c *Conn) close() error {
	if r := recover(); r != nil {
		defer log.ErrLogger.Debug().Str("context", "conn.closing").Msgf("panic recovered '%v'", debug.Stack())
	}

	// Unsubscribe from everything, no need to lock since each Unsubscribe is
	// already locked. Locking the 'Close()' would result in a deadlock.
	// Don't close clustered connection, their servers are not being shut down.
	if c.clnode == nil {
		for _, stat := range c.subs.All() {
			store.Connection.Delete(c.clientid.Contract(), stat.Topic, stat.ID)
			// Decrement the subscription counter
			c.service.meter.Subscriptions.Dec(1)
		}
	}

	Globals.ConnCache.Delete(c.connid)
	defer log.ConnLogger.Info().Str("context", "conn.close").Int64("connid", int64(c.connid)).Msg("conn closed")
	Globals.Cluster.connGone(c)
	// Decrement the connection counter
	c.service.meter.Connections.Dec(1)
	return c.socket.Close()
}
