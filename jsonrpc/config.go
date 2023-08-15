package jsonrpc

import (
	"time"
)

// Config represents the configuration of the json rpc
type Config struct {
	// Host defines the network adapter that will be used to serve the HTTP requests
	Host string `mapstructure:"Host"`

	// Port defines the port to serve the endpoints via HTTP
	Port int `mapstructure:"Port"`

	// ReadTimeout is the HTTP server read timeout
	// check net/http.server.ReadTimeout and net/http.server.ReadHeaderTimeout
	ReadTimeout Duration `mapstructure:"ReadTimeout"`

	// WriteTimeout is the HTTP server write timeout
	// check net/http.server.WriteTimeout
	WriteTimeout Duration `mapstructure:"WriteTimeout"`

	// MaxRequestsPerIPAndSecond defines how much requests a single IP can
	// send within a single second
	MaxRequestsPerIPAndSecond float64 `mapstructure:"MaxRequestsPerIPAndSecond"`

	// SequencerNodeURI is used allow Non-Sequencer nodes
	// to relay transactions to the Sequencer node
	SequencerNodeURI string `mapstructure:"SequencerNodeURI"`

	// MaxCumulativeGasUsed is the max gas allowed per batch
	MaxCumulativeGasUsed uint64

	// WebSockets configuration
	WebSockets WebSocketsConfig `mapstructure:"WebSockets"`

	// EnableL2SuggestedGasPricePolling enables polling of the L2 gas price to block tx in the RPC with lower gas price.
	// EnableL2SuggestedGasPricePolling bool `mapstructure:"EnableL2SuggestedGasPricePolling"`

	// TraceBatchUseHTTPS enables, in the debug_traceBatchByNum endpoint, the use of the HTTPS protocol (instead of HTTP)
	// to do the parallel requests to RPC.debug_traceTransaction endpoint
	TraceBatchUseHTTPS bool `mapstructure:"TraceBatchUseHTTPS"`
}

// WebSocketsConfig has parameters to config the rpc websocket support
type WebSocketsConfig struct {
	// Enabled defines if the WebSocket requests are enabled or disabled
	Enabled bool `mapstructure:"Enabled"`

	// Host defines the network adapter that will be used to serve the WS requests
	Host string `mapstructure:"Host"`

	// Port defines the port to serve the endpoints via WS
	Port int `mapstructure:"Port"`
}

// Duration is a wrapper type that parses time duration from text.
type Duration struct {
	time.Duration `validate:"required"`
}

// UnmarshalText unmarshalls time duration from text.
func (d *Duration) UnmarshalText(data []byte) error {
	duration, err := time.ParseDuration(string(data))
	if err != nil {
		return err
	}
	d.Duration = duration
	return nil
}

// NewDuration returns Duration wrapper
func NewDuration(duration time.Duration) Duration {
	return Duration{time.Duration(duration)}
}
