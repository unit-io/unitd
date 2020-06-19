package broker

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/saffat-in/trace/websocket"
	"github.com/unit-io/unitd/config"
	lp "github.com/unit-io/unitd/net"
	"github.com/unit-io/unitd/net/listener"
	"github.com/unit-io/unitd/pkg/crypto"
	"github.com/unit-io/unitd/pkg/log"
	"github.com/unit-io/unitd/pkg/stats"
	"github.com/unit-io/unitd/pkg/uid"

	// Database store
	_ "github.com/unit-io/unitd/db/unitdb"
	"github.com/unit-io/unitd/store"
)

//Service is a main struct
type Service struct {
	PID     uint32             // The processid is unique Id for the application
	MAC     *crypto.MAC        // The MAC to use for decoding and encoding keys.
	cache   *sync.Map          // The cache for the contracts.
	context context.Context    // context for the service
	config  *config.Config     // The configuration for the service.
	cancel  context.CancelFunc // cancellation function
	start   time.Time          // The service start time
	http    *http.Server       // The underlying HTTP server.
	tcp     *lp.Server         // The underlying TCP server.
	grpc    *lp.Server         // The underlying GRPC server.
	meter   *Meter             // The metircs to measure timeseries on message events
	stats   *stats.Stats
}

func NewService(ctx context.Context, cfg *config.Config) (s *Service, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	s = &Service{
		PID:     uid.NewUnique(),
		cache:   new(sync.Map),
		context: ctx,
		config:  cfg,
		cancel:  cancel,
		start:   time.Now(),
		// subscriptions: message.NewSubscriptions(),
		http:  new(http.Server),
		tcp:   new(lp.Server),
		grpc:  new(lp.Server),
		meter: NewMeter(),
		stats: stats.New(&stats.Config{Addr: "localhost:8094", Size: 50}, stats.MaxPacketSize(1400), stats.MetricPrefix("trace")),
	}

	// Create a new HTTP request multiplexer
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.onRequest)

	// Varz
	if cfg.VarzPath != "" {
		mux.HandleFunc(cfg.VarzPath, s.HandleVarz)
		log.Info("service", "Stats variables exposed at "+cfg.VarzPath)
	}

	//attach handlers
	s.grpc.StreamHandler = s.onAcceptConn
	s.http.Handler = mux
	s.tcp.TcpHandler = s.onAcceptConn

	// Create a new MAC from the key.
	if s.MAC, err = crypto.New([]byte(s.config.Encryption(s.config.EncryptionConfig).Key)); err != nil {
		return nil, err
	}

	// Open database connection
	err = store.Open(string(s.config.Store))
	if err != nil {
		log.Fatal("service", "Failed to connect to DB:", err)
	}

	return s, nil
}

//Listen starts the service
func (s *Service) Listen() (err error) {
	defer s.Close()
	s.hookSignals()

	s.listen(s.config.Listen)

	log.Info("service", "service started")
	select {}
}

//listen configures main listerner on specefied address
func (s *Service) listen(addr string) {

	//Create a new listener
	log.Info("service.listen", "starting the listner at "+addr)

	l, err := listener.New(addr)
	if err != nil {
		panic(err)
	}

	l.SetReadTimeout(120 * time.Second)

	// Configure the protos
	s.grpc.ListenAndServe(s.config.GrpcListen, false, nil)
	l.ServeCallback(listener.MatchWS("GET"), s.http.Serve)
	l.ServeCallback(listener.MatchAny(), s.tcp.Serve)

	go l.Serve()
}

// Handle a new connection request
func (s *Service) onAcceptConn(t net.Conn, proto lp.Proto) {
	conn := s.newConn(t, proto)
	go conn.readLoop()
	go conn.writeLoop()
}

// Handle a new HTTP request.
func (s *Service) onRequest(w http.ResponseWriter, r *http.Request) {
	if ws, ok := websocket.Handler(w, r); ok {
		s.onAcceptConn(ws, lp.WEBSOCK)
		return
	}
}

func (s *Service) onSignal(sig os.Signal) {
	switch sig {
	case syscall.SIGTERM:
		fallthrough
	case syscall.SIGINT:
		log.Info("service.onSignal", "received signal, exiting..."+sig.String())
		s.Close()
		os.Exit(0)
	}
}

func (s *Service) hookSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for sig := range c {
			s.onSignal(sig)
		}
	}()
}

func (s *Service) Close() {
	if s.cancel != nil {
		s.cancel()
	}

	s.meter.UnregisterAll()
	s.stats.Unregister()

	store.Close()

	// Shutdown local cluster node, if it's a part of a cluster.
	Globals.Cluster.shutdown()
}
