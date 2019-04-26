package trace

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"
	"encoding/json"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"

	traced "github.com/tracedb/trace/broker"
)

type Stats struct {
	Server          string
	ResponseTimeout		internal.Duration
	Interval			internal.Duration
	FlushInterval		internal.Duration

	client *http.Client
}

var sampleConfig = `
	## The address of the monitoring endpoint of the trace broker
  	server = "http://localhost:6060"
	## Maximum time to receive response
	#response_timeout = "5s"
	#interval = "1m"
	#metric_batch_size = 5
	#flush_interval = "1m"
`

func (s *Stats) SampleConfig() string {
	return sampleConfig
}

func (s *Stats) Description() string {
	return "Provides metrics about the state of a Trace server"
}

func (s *Stats) Gather(acc telegraf.Accumulator) error {
	url, err := url.Parse(s.Server)
	if err != nil {
		return err
	}
	url.Path = path.Join(url.Path, "varz")

	if s.client == nil {
		s.client = s.createHTTPClient()
	}
	resp, err := s.client.Get(url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	stats := new(traced.Varz)
	err = json.Unmarshal([]byte(bytes), &stats)
	if err != nil {
		return err
	}
	acc.AddFields("trace",
		map[string]interface{}{
			"Uptime": stats.Uptime,
			"Connections": stats.Connections,
			"InMsgs": stats.InMsgs,
			"OutMsgs": stats.OutMsgs,
			"InBytes": stats.InBytes,
			"OutBytes": stats.OutBytes,
			"Subscriptions": stats.Subscriptions,
			"HMean": stats.HMean,
			"P50": stats.P50,
			"P75": stats.P75,
			"P95": stats.P95,
			"P99": stats.P99,
			"P999": stats.P999,
			"Long5p": stats.Long5p,
			"Short5p": stats.Short5p,
			"Max": stats.Max,
			"Min": stats.Min,
			"StdDev": stats.StdDev,
		},
		map[string]string{"stats": "conn_traffic"},
		time.Now())

	return nil
}

func (s *Stats) createHTTPClient() *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	timeout := s.ResponseTimeout.Duration
	if timeout == time.Duration(0) {
	 	timeout = 5 * time.Second
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

func init() {
	inputs.Add("trace", func() telegraf.Input {
		return &Stats{
			Server: "http://localhost:6060",
		}
	})
}