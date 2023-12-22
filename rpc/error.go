package rpc

import (
	"errors"
	"fmt"

	"github.com/0xPolygon/cdk-data-availability/types"
)

const (
	// DefaultErrorCode rpc default error code
	DefaultErrorCode = -32000
	// RevertedErrorCode error code for reverted txs
	RevertedErrorCode = 3
	// InvalidRequestErrorCode error code for invalid requests
	InvalidRequestErrorCode = -32600
	// NotFoundErrorCode error code for not found objects
	NotFoundErrorCode = -32601
	// InvalidParamsErrorCode error code for invalid parameters
	InvalidParamsErrorCode = -32602
	// ParserErrorCode error code for parsing errors
	ParserErrorCode = -32700
	// AccessDeniedCode error code when requests are denied
	AccessDeniedCode = -32800
)

var (
	// invalidJSONReqErr denotes error that is returned when invalid JSON request is received
	invalidJSONReqErr = errors.New("Invalid json request")
)

// Error interface
type Error interface {
	Error() string
	ErrorCode() int
	ErrorData() *[]byte
}

// ErrorObject is a jsonrpc error
type ErrorObject struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    *types.ArgBytes `json:"data,omitempty"`
}

// RPCError represents an error returned by a JSON RPC endpoint.
type RPCError struct {
	err  string
	code int
	data *[]byte
}

// NewRPCError creates a new error instance to be returned by the RPC endpoints
func NewRPCError(code int, err string, args ...interface{}) *RPCError {
	return NewRPCErrorWithData(code, err, nil, args...)
}

// NewRPCErrorWithData creates a new error instance with data to be returned by the RPC endpoints
func NewRPCErrorWithData(code int, err string, data *[]byte, args ...interface{}) *RPCError {
	var errMessage string
	if len(args) > 0 {
		errMessage = fmt.Sprintf(err, args...)
	} else {
		errMessage = err
	}
	return &RPCError{code: code, err: errMessage, data: data}
}

// Error returns the error message.
func (e *RPCError) Error() string {
	return e.err
}

// ErrorCode returns the error code.
func (e *RPCError) ErrorCode() int {
	return e.code
}

// ErrorData returns the error data.
func (e *RPCError) ErrorData() *[]byte {
	return e.data
}
