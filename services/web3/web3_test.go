package web3

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEndpoints_ClientVersion(t *testing.T) {
	e := Endpoints{}

	clientVersion, err := e.ClientVersion()
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(
		"cdk-dac/v0.1.0/%s-%s/%s",
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version(),
	), clientVersion)
}
