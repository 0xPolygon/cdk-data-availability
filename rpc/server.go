package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/didip/tollbooth/v6"
)

// Server is an API backend to handle RPC requests
type Server struct {
	config  Config
	handler *Handler
	srv     *http.Server
}

// Service implementation of a service an it's name
type Service struct {
	Name    string
	Service interface{}
}

// NewServer returns the JsonRPC server
func NewServer(
	cfg Config,
	services []Service,
) *Server {
	handler := newJSONRpcHandler()

	for _, service := range services {
		handler.registerService(service)
	}

	srv := &Server{
		config:  cfg,
		handler: handler,
	}
	return srv
}

// Start initializes the JSON RPC server to listen for request
func (s *Server) Start() error {
	return s.startHTTP()
}

// startHTTP starts a server to respond http requests
func (s *Server) startHTTP() error {
	if s.srv != nil {
		return fmt.Errorf("server already started")
	}

	address := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Errorf("failed to create tcp listener: %v", err)
		return err
	}

	mux := http.NewServeMux()

	lmt := tollbooth.NewLimiter(s.config.MaxRequestsPerIPAndSecond, nil)
	mux.Handle("/", tollbooth.LimitFuncHandler(lmt, s.handle))

	s.srv = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: s.config.ReadTimeout.Duration,
		ReadTimeout:       s.config.ReadTimeout.Duration,
		WriteTimeout:      s.config.WriteTimeout.Duration,
	}
	log.Infof("http server started: %s", address)
	if err := s.srv.Serve(lis); err != nil {
		if err == http.ErrServerClosed {
			log.Infof("http server stopped")
			return nil
		}
		log.Errorf("closed http connection: %v", err)
		return err
	}
	return nil
}

// Stop shutdown the rpc server
func (s *Server) Stop() error {
	if s.srv != nil {
		if err := s.srv.Shutdown(context.Background()); err != nil {
			return err
		}

		if err := s.srv.Close(); err != nil {
			return err
		}
		s.srv = nil
	}

	return nil
}

func (s *Server) handle(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if (*req).Method == "OPTIONS" {
		// TODO(pg): need to count it in the metrics?
		return
	}

	if req.Method == "GET" {
		// TODO(pg): need to count it in the metrics?
		_, err := w.Write([]byte("zkEVM JSON RPC Server"))
		if err != nil {
			log.Error(err)
		}
		return
	}

	if req.Method != "POST" {
		err := errors.New("method " + req.Method + " not allowed")
		s.handleInvalidRequest(w, err)
		return
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		s.handleInvalidRequest(w, err)
		return
	}

	single, err := s.isSingleRequest(data)
	if err != nil {
		s.handleInvalidRequest(w, err)
		return
	}

	start := time.Now()
	var respLen int
	if single {
		respLen = s.handleSingleRequest(req, w, data)
	} else {
		respLen = s.handleBatchRequest(req, w, data)
	}
	combinedLog(req, start, http.StatusOK, respLen)
}

func (s *Server) isSingleRequest(data []byte) (bool, Error) {
	x := bytes.TrimLeft(data, " \t\r\n")

	if len(x) == 0 {
		return false, NewRPCError(InvalidRequestErrorCode, invalidJSONReqErr.Error())
	}

	return x[0] == '{', nil
}

func (s *Server) handleSingleRequest(httpRequest *http.Request, w http.ResponseWriter, data []byte) int {
	request, err := s.parseRequest(data)
	if err != nil {
		handleError(w, err)
		return 0
	}
	req := handleRequest{Request: request, HttpRequest: httpRequest}
	response := s.handler.Handle(req)

	respBytes, err := json.Marshal(response)
	if err != nil {
		handleError(w, err)
		return 0
	}

	_, err = w.Write(respBytes)
	if err != nil {
		handleError(w, err)
		return 0
	}
	return len(respBytes)
}

func (s *Server) handleBatchRequest(httpRequest *http.Request, w http.ResponseWriter, data []byte) int {
	requests, err := s.parseRequests(data)
	if err != nil {
		handleError(w, err)
		return 0
	}

	responses := make([]Response, 0, len(requests))

	for _, request := range requests {
		req := handleRequest{Request: request, HttpRequest: httpRequest}
		response := s.handler.Handle(req)
		responses = append(responses, response)
	}

	respBytes, _ := json.Marshal(responses)
	_, err = w.Write(respBytes)
	if err != nil {
		log.Error(err)
		return 0
	}
	return len(respBytes)
}

func (s *Server) parseRequest(data []byte) (Request, error) {
	var req Request

	if err := json.Unmarshal(data, &req); err != nil {
		return Request{}, NewRPCError(InvalidRequestErrorCode, invalidJSONReqErr.Error())
	}

	return req, nil
}

func (s *Server) parseRequests(data []byte) ([]Request, error) {
	var requests []Request

	if err := json.Unmarshal(data, &requests); err != nil {
		return nil, NewRPCError(InvalidRequestErrorCode, invalidJSONReqErr.Error())
	}

	return requests, nil
}

func (s *Server) handleInvalidRequest(w http.ResponseWriter, err error) {
	handleError(w, err)
}

func handleError(w http.ResponseWriter, err error) {
	log.Error(err)
	w.WriteHeader(http.StatusInternalServerError)
	_, err = w.Write([]byte(err.Error()))
	if err != nil {
		log.Error(err)
	}
}

// RPCErrorResponse formats error to be returned through RPC
func RPCErrorResponse(code int, message string, err error) (interface{}, Error) {
	return RPCErrorResponseWithData(code, message, nil, err)
}

// RPCErrorResponseWithData formats error to be returned through RPC
func RPCErrorResponseWithData(code int, message string, data *[]byte, err error) (interface{}, Error) {
	if err != nil {
		log.Errorf("%v: %v", message, err.Error())
	} else {
		log.Error(message)
	}
	return nil, NewRPCErrorWithData(code, message, data)
}

func combinedLog(r *http.Request, start time.Time, httpStatus, dataLen int) {
	log.Infof("%s - - %s \"%s %s %s\" %d %d \"%s\" \"%s\"",
		r.RemoteAddr,
		start.Format("[02/Jan/2006:15:04:05 -0700]"),
		r.Method,
		r.URL.Path,
		r.Proto,
		httpStatus,
		dataLen,
		r.Host,
		r.UserAgent(),
	)
}
