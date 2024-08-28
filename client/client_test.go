package client

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestClient_GetStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		result     string
		status     *types.DACStatus
		statusCode int
		err        error
	}{
		{
			name:   "successfully got status",
			result: `{"result":{"version":"v1.0.0","uptime":"123","key_count":2,"backfill_progress":5}}`,
			status: &types.DACStatus{
				Uptime:           "123",
				Version:          "v1.0.0",
				KeyCount:         2,
				BackfillProgress: 5,
			},
		},
		{
			name:   "error returned by server",
			result: `{"error":{"code":123,"message":"test error"}}`,
			err:    errors.New("123 test error"),
		},
		{
			name:       "unsuccessful status code returned by server",
			statusCode: http.StatusInternalServerError,
			err:        errors.New("invalid status code, expected: 200, found: 500"),
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var res rpc.Request
				require.NoError(t, json.NewDecoder(r.Body).Decode(&res))
				require.Equal(t, "status_getStatus", res.Method)

				if tt.statusCode > 0 {
					w.WriteHeader(tt.statusCode)
				}

				_, err := fmt.Fprint(w, tt.result)
				require.NoError(t, err)
			}))
			defer srv.Close()

			client := New(srv.URL)

			got, err := client.GetStatus(context.Background())
			if tt.err != nil {
				require.Error(t, err)
				require.EqualError(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.status, got)
			}
		})
	}
}

func TestClient_SignSequence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ss         types.SignedSequence
		result     string
		signature  []byte
		statusCode int
		err        error
	}{
		{
			name: "successfully signed sequence",
			ss: types.SignedSequence{
				Sequence:  types.Sequence{},
				Signature: []byte("signature0"),
			},
			result:    fmt.Sprintf(`{"result":"%s"}`, hex.EncodeToString([]byte("signature1"))),
			signature: []byte("signature1"),
		},
		{
			name: "error returned by server",
			ss: types.SignedSequence{
				Sequence:  types.Sequence{},
				Signature: []byte("signature0"),
			},
			result: `{"error":{"code":123,"message":"test error"}}`,
			err:    errors.New("123 test error"),
		},
		{
			name: "invalid signature returned by server",
			ss: types.SignedSequence{
				Sequence:  types.Sequence{},
				Signature: []byte("signature0"),
			},
			result: `{"result":"invalid-signature"}`,
		},
		{
			name: "unsuccessful status code returned by server",
			ss: types.SignedSequence{
				Sequence:  types.Sequence{},
				Signature: []byte("signature0"),
			},
			statusCode: http.StatusInternalServerError,
			err:        errors.New("invalid status code, expected: 200, found: 500"),
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var res rpc.Request
				require.NoError(t, json.NewDecoder(r.Body).Decode(&res))
				require.Equal(t, "datacom_signSequence", res.Method)

				var params []types.SignedSequence
				require.NoError(t, json.Unmarshal(res.Params, &params))
				require.Equal(t, tt.ss, params[0])

				if tt.statusCode > 0 {
					w.WriteHeader(tt.statusCode)
				}

				_, err := fmt.Fprint(w, tt.result)
				require.NoError(t, err)
			}))
			defer srv.Close()

			client := New(srv.URL)

			got, err := client.SignSequence(context.Background(), tt.ss)
			if tt.err != nil {
				require.Error(t, err)
				require.EqualError(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.signature, got)
			}
		})
	}
}

func TestClient_GetOffChainData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		hash       common.Hash
		result     string
		data       []byte
		statusCode int
		err        error
	}{
		{
			name:   "successfully got offhcain data",
			hash:   common.BytesToHash([]byte("hash")),
			result: fmt.Sprintf(`{"result":"%s"}`, hex.EncodeToString([]byte("offchaindata"))),
			data:   []byte("offchaindata"),
		},
		{
			name:   "error returned by server",
			hash:   common.BytesToHash([]byte("hash")),
			result: `{"error":{"code":123,"message":"test error"}}`,
			err:    errors.New("123 test error"),
		},
		{
			name:   "invalid offchain data returned by server",
			hash:   common.BytesToHash([]byte("hash")),
			result: `{"result":"invalid-signature"}`,
		},
		{
			name:       "unsuccessful status code returned by server",
			hash:       common.BytesToHash([]byte("hash")),
			statusCode: http.StatusUnauthorized,
			err:        errors.New("invalid status code, expected: 200, found: 401"),
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var res rpc.Request
				require.NoError(t, json.NewDecoder(r.Body).Decode(&res))
				require.Equal(t, "sync_getOffChainData", res.Method)

				var params []common.Hash
				require.NoError(t, json.Unmarshal(res.Params, &params))
				require.Equal(t, tt.hash, params[0])

				if tt.statusCode > 0 {
					w.WriteHeader(tt.statusCode)
				}

				_, err := fmt.Fprint(w, tt.result)
				require.NoError(t, err)
			}))
			defer svr.Close()

			c := &client{url: svr.URL}

			got, err := c.GetOffChainData(context.Background(), tt.hash)
			if tt.err != nil {
				require.Error(t, err)
				require.EqualError(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.data, got)
			}
		})
	}
}

func TestClient_ListOffChainData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		hashes     []common.Hash
		result     string
		data       map[common.Hash][]byte
		statusCode int
		err        error
	}{
		{
			name:   "successfully got offhcain data",
			hashes: []common.Hash{common.BytesToHash([]byte("hash"))},
			result: fmt.Sprintf(`{"result":{"%s":"%s"}}`,
				common.BytesToHash([]byte("hash")).Hex(), hex.EncodeToString([]byte("offchaindata"))),
			data: map[common.Hash][]byte{
				common.BytesToHash([]byte("hash")): []byte("offchaindata"),
			},
		},
		{
			name:   "error returned by server",
			hashes: []common.Hash{common.BytesToHash([]byte("hash"))},
			result: `{"error":{"code":123,"message":"test error"}}`,
			err:    errors.New("123 test error"),
		},
		{
			name:   "invalid offchain data returned by server",
			hashes: []common.Hash{common.BytesToHash([]byte("hash"))},
			result: fmt.Sprintf(`{"result":{"%s":"invalid-signature"}}`,
				common.BytesToHash([]byte("hash")).Hex()),
			data: map[common.Hash][]byte{
				common.BytesToHash([]byte("hash")): nil,
			},
		},
		{
			name:       "unsuccessful status code returned by server",
			hashes:     []common.Hash{common.BytesToHash([]byte("hash"))},
			statusCode: http.StatusUnauthorized,
			err:        errors.New("invalid status code, expected: 200, found: 401"),
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var res rpc.Request
				require.NoError(t, json.NewDecoder(r.Body).Decode(&res))
				require.Equal(t, "sync_listOffChainData", res.Method)

				var params [][]common.Hash
				require.NoError(t, json.Unmarshal(res.Params, &params))
				require.Equal(t, tt.hashes, params[0])

				if tt.statusCode > 0 {
					w.WriteHeader(tt.statusCode)
				}

				_, err := fmt.Fprint(w, tt.result)
				require.NoError(t, err)
			}))
			defer svr.Close()

			c := &client{url: svr.URL}

			got, err := c.ListOffChainData(context.Background(), tt.hashes)
			if tt.err != nil {
				require.Error(t, err)
				require.EqualError(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.data, got)
			}
		})
	}
}

func TestClient_SignSequenceBanana(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ssb        types.SignedSequenceBanana
		result     string
		signature  []byte
		statusCode int
		err        error
	}{
		{
			name: "successfully signed banana sequence",
			ssb: types.SignedSequenceBanana{
				Sequence:  types.SequenceBanana{},
				Signature: []byte("signature00"),
			},
			result:    fmt.Sprintf(`{"result":"%s"}`, hex.EncodeToString([]byte("signature11"))),
			signature: []byte("signature11"),
		},
		{
			name: "error returned by rpc server",
			ssb: types.SignedSequenceBanana{
				Sequence:  types.SequenceBanana{},
				Signature: []byte("signature00"),
			},
			result: `{"error":{"code":123,"message":"test error"}}`,
			err:    errors.New("123 test error"),
		},
		{
			name: "invalid signature returned by rpc server",
			ssb: types.SignedSequenceBanana{
				Sequence:  types.SequenceBanana{},
				Signature: []byte("signature00"),
			},
			result: `{"result":"invalid-signature"}`,
		},
		{
			name: "unsuccessful status code returned by rpc server",
			ssb: types.SignedSequenceBanana{
				Sequence:  types.SequenceBanana{},
				Signature: []byte("signature00"),
			},
			statusCode: http.StatusInternalServerError,
			err:        errors.New("invalid status code, expected: 200, found: 500"),
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request rpc.Request
				require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
				require.Equal(t, "datacom_signSequenceBanana", request.Method)

				var params []types.SignedSequenceBanana
				require.NoError(t, json.Unmarshal(request.Params, &params))
				require.Equal(t, tt.ssb, params[0])

				if tt.statusCode > 0 {
					w.WriteHeader(tt.statusCode)
				}

				_, err := fmt.Fprint(w, tt.result)
				require.NoError(t, err)
			}))
			defer srv.Close()

			client := New(srv.URL)

			result, err := client.SignSequenceBanana(context.Background(), tt.ssb)
			if tt.err != nil {
				require.Error(t, err)
				require.EqualError(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.signature, result)
			}
		})
	}
}

func TestClient_Factory_New(t *testing.T) {
	t.Parallel()

	url := "http://example.com"
	f := NewFactory()

	c := f.New(url)
	require.NotNil(t, c)

	client, ok := c.(*client)
	require.True(t, ok)
	require.Equal(t, url, client.url)
}
