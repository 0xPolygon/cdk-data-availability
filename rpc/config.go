package rpc

import "github.com/0xPolygon/cdk-data-availability/config/types"

// Config represents the configuration of the json rpc
type Config struct {
	// Host defines the network adapter that will be used to serve the HTTP requests
	Host string `mapstructure:"Host"`

	// Port defines the port to serve the endpoints via HTTP
	Port int `mapstructure:"Port"`

	// ReadTimeout is the HTTP server read timeout
	// check net/http.server.ReadTimeout and net/http.server.ReadHeaderTimeout
	ReadTimeout types.Duration `mapstructure:"ReadTimeout"`

	// WriteTimeout is the HTTP server write timeout
	// check net/http.server.WriteTimeout
	WriteTimeout types.Duration `mapstructure:"WriteTimeout"`

	// MaxRequestsPerIPAndSecond defines how much requests a single IP can
	// send within a single second
	MaxRequestsPerIPAndSecond float64 `mapstructure:"MaxRequestsPerIPAndSecond"`
}
