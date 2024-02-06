package rpc

import (
	"encoding/json"

	"github.com/0xPolygon/cdk-data-availability/types"
)

// Request is a jsonrpc request
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response is a jsonrpc  success response
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *ErrorObject    `json:"error"`
}

// NewResponse returns Success/Error response object
func NewResponse(req Request, reply []byte, err Error) Response {
	var result json.RawMessage
	if reply != nil {
		result = reply
	}

	var errorObj *ErrorObject
	if err != nil {
		errorObj = &ErrorObject{
			Code:    err.ErrorCode(),
			Message: err.Error(),
		}
		if err.ErrorData() != nil {
			errorObj.Data = types.ArgBytesPtr(*err.ErrorData())
		}
	}

	return Response{
		JSONRPC: req.JSONRPC,
		ID:      req.ID,
		Result:  result,
		Error:   errorObj,
	}
}

// Bytes return the serialized response
func (s Response) Bytes() ([]byte, error) {
	return json.Marshal(s)
}
