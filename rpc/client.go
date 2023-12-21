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
	return JSONRPCCallWithContext(context.Background(), url, method, params)
}

// JSONRPCCallWithContext executes a 2.0 JSON RPC HTTP Post Request to the provided URL with
// the provided method and parameters, which is compatible with the Ethereum
// JSON RPC Server.
func JSONRPCCallWithContext(ctx context.Context, url, method string, parameters ...interface{}) (Response, error) {
	params, err := json.Marshal(parameters)
	if err != nil {
		return Response{}, err
	}

	req := Request{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  method,
		Params:  params,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return Response{}, err
	}

	reqBodyReader := bytes.NewReader(reqBody)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reqBodyReader)
	if err != nil {
		return Response{}, err
	}

	httpReq.Header.Add("Content-type", "application/json")

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
