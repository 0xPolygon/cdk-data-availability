package types

import (
	"encoding/json"
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
	JSONRPC string
	ID      interface{}
	Result  json.RawMessage
	Error   *ErrorObject
}

// ErrorObject is a jsonrpc error
type ErrorObject struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    *ArgBytes `json:"data,omitempty"`
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
			errorObj.Data = ArgBytesPtr(*err.ErrorData())
		}
	}

	return Response{
		JSONRPC: req.JSONRPC,
		ID:      req.ID,
		Result:  result,
		Error:   errorObj,
	}
}

// MarshalJSON customizes the JSON representation of the response.
func (r Response) MarshalJSON() ([]byte, error) {
	if r.Error != nil {
		return json.Marshal(struct {
			JSONRPC string       `json:"jsonrpc"`
			ID      interface{}  `json:"id"`
			Error   *ErrorObject `json:"error"`
		}{
			JSONRPC: r.JSONRPC,
			ID:      r.ID,
			Error:   r.Error,
		})
	}

	return json.Marshal(struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      interface{}     `json:"id"`
		Result  json.RawMessage `json:"result"`
	}{
		JSONRPC: r.JSONRPC,
		ID:      r.ID,
		Result:  r.Result,
	})
}

// Bytes return the serialized response
func (s Response) Bytes() ([]byte, error) {
	return json.Marshal(s)
}
