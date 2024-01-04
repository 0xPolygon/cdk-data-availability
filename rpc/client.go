package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// JSONRPCCall calls JSONRPCCallWithContext with the default context
func JSONRPCCall(url, method string, params ...interface{}) (Response, error) {
	return JSONRPCCallWithContext(context.Background(), url, method, params...)
}

// JSONRPCCallWithContext executes a 2.0 JSON RPC HTTP Post Request to the provided URL with
// the provided method and parameters, which is compatible with the Ethereum
// JSON RPC Server.
func JSONRPCCallWithContext(ctx context.Context, url, method string, parameters ...interface{}) (Response, error) {
	httpReq, err := BuildJsonHTTPRequest(ctx, url, method, parameters...)
	if err != nil {
		return Response{}, err
	}

	httpRes, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return Response{}, err
	}

	if httpRes.Body != nil {
		defer httpRes.Body.Close()
	}

	if httpRes.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("invalid status code, expected: %v, found: %v", http.StatusOK, httpRes.StatusCode)
	}

	var res Response
	if err = json.NewDecoder(httpRes.Body).Decode(&res); err != nil {
		return Response{}, err
	}

	return res, nil
}

// BuildJsonHTTPRequest creates JSON RPC http request using provided url, method and parameters
func BuildJsonHTTPRequest(ctx context.Context, url, method string, parameters ...interface{}) (*http.Request, error) {
	params, err := json.Marshal(parameters)
	if err != nil {
		return nil, err
	}

	req := Request{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  method,
		Params:  params,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	return BuildJsonHttpRequestWithBody(ctx, url, reqBody)
}

// BuildJsonHttpRequestWithBody creates JSON RPC http request using provided url and request body
func BuildJsonHttpRequestWithBody(ctx context.Context, url string, reqBody []byte) (*http.Request, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Add("Content-type", "application/json")

	return httpReq, nil
}
