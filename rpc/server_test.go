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
	const (
		funcName   = "greeter_handleReq"
		paramValue = "John Doe"
	)

	// Create a test Server with mock configuration and service
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
	url := fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port)

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

	expectedResponse := fmt.Sprintf("\"Hello, %s!\"", paramValue)

	t.Run("handle single request", func(t *testing.T) {
		// Create a new request with the specified method, URL, and payload
		req, err := BuildJsonHTTPRequest(context.Background(), url, funcName, paramValue)
		require.NoError(t, err)

		// Perform an HTTP request to the server
		respRecorder := httptest.NewRecorder()
		server.handle(respRecorder, req)

		// Assert the response
		require.Equal(t, http.StatusOK, respRecorder.Code)
		var resp Response
		err = json.Unmarshal(respRecorder.Body.Bytes(), &resp)
		require.NoError(t, err)

		require.Equal(t, expectedResponse, string(resp.Result))
	})

	t.Run("handle batch request", func(t *testing.T) {
		const batchSize = 3

		params, err := json.Marshal([]interface{}{paramValue})
		require.NoError(t, err)

		// Construct batch request
		reqs := make([]Request, batchSize)
		for i := 0; i < batchSize; i++ {
			reqs[i] = Request{
				JSONRPC: "2.0",
				ID:      float64(i + 1),
				Method:  funcName,
				Params:  params,
			}
		}

		reqBody, err := json.Marshal(reqs)
		require.NoError(t, err)

		httpReq, err := BuildJsonHttpRequestWithBody(context.Background(), url, reqBody)
		require.NoError(t, err)

		respRecorder := httptest.NewRecorder()
		server.handle(respRecorder, httpReq)

		require.Equal(t, http.StatusOK, respRecorder.Code)

		// Parse the response body
		var resp []Response
		err = json.Unmarshal(respRecorder.Body.Bytes(), &resp)
		require.NoError(t, err)

		require.Len(t, resp, batchSize)
		for i := 0; i < batchSize; i++ {
			require.Equal(t, float64(i+1), resp[i].ID)
			require.Equal(t, expectedResponse, string(resp[i].Result))
		}
	})
}

type greeterService struct{}

// Mock implementation of a service method
func (s *greeterService) HandleReq(name string) (interface{}, Error) {
	return fmt.Sprintf("Hello, %s!", name), nil
}
