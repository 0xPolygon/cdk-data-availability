package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-data-availability/config/types"
	"github.com/stretchr/testify/require"
)

func Test_ServerHandleRequest(t *testing.T) {
	// Create a test Server with mock configuration and services
	cfg := Config{
		Host:                      "localhost",
		Port:                      8080,
		MaxRequestsPerIPAndSecond: 100,
		ReadTimeout:               types.Duration{Duration: 10},
		WriteTimeout:              types.Duration{Duration: 10},
	}
	services := []Service{
		{Name: "greeter", Service: &greeterService{}},
	}
	server := NewServer(cfg, services)

	defer func() {
		// Stop the server
		err := server.Stop()
		require.NoError(t, err)
	}()

	// Start the server
	go func() {
		err := server.Start()
		require.NoError(t, err)
	}()

	// Allow some time for the server to start
	// You might need to adjust the sleep duration based on your system and test conditions
	<-time.After(100 * time.Millisecond)

	t.Run("handle single request", func(t *testing.T) {
		url := fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port)
		name := "John Doe"

		// Create a new request with the specified method, URL, and payload
		req, err := BuildJsonHTTPRequest(context.Background(), url, "greeter_handleReq", name)
		require.NoError(t, err)

		// // Perform an HTTP request to the server (you can use an HTTP client library or http.NewRequest)
		respRecorder := httptest.NewRecorder()
		server.handle(respRecorder, req)

		// // Assert the response (you might need to adjust the assertions based on your actual implementation)
		require.Equal(t, http.StatusOK, respRecorder.Code)
		var resp Response
		err = json.Unmarshal(respRecorder.Body.Bytes(), &resp)
		require.NoError(t, err)

		expectedResponse := fmt.Sprintf("\"Hello, %s!\"", name)
		require.Equal(t, expectedResponse, string(resp.Result))
	})
}

type greeterService struct{}

// Mock implementation of a service method
func (s *greeterService) HandleReq(name string) (interface{}, Error) {
	return fmt.Sprintf("Hello, %s!", name), nil
}
