package web3

import (
	"fmt"
	"runtime"

	dataavailability "github.com/0xPolygon/cdk-data-availability"
	"github.com/0xPolygon/cdk-data-availability/rpc"
)

const (
	// APIWEB3 is the namespace of the web3 service
	APIWEB3 = "web3"

	clientVersionTemplate = "cdk-dac/%s/%s-%s/%s"
)

// Endpoints contains implementations for the "zkevm" RPC endpoints
type Endpoints struct {
}

// NewEndpoints returns Endpoints
func NewEndpoints() *Endpoints {
	return &Endpoints{}
}

// ClientVersion returns the client version info
func (e *Endpoints) ClientVersion() (interface{}, rpc.Error) {
	var version string
	if dataavailability.Version != "" {
		version = dataavailability.Version
	} else if dataavailability.GitRev != "" {
		version = dataavailability.GitRev[:8]
	}

	return fmt.Sprintf(
		clientVersionTemplate,
		version,
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version(),
	), nil
}
