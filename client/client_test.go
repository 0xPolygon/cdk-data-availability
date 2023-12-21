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

	"github.com/ethereum/go-ethereum/common"

	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/stretchr/testify/require"
)

func TestClient_SignSequence(t *testing.T) {
	tests := []struct {
		name      string
		ss        types.SignedSequence
		result    string
		signature []byte
		err       error
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
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var res rpc.Request
				require.NoError(t, json.NewDecoder(r.Body).Decode(&res))
				require.Equal(t, "datacom_signSequence", res.Method)

				var params []types.SignedSequence
				require.NoError(t, json.Unmarshal(res.Params, &params))
				require.Equal(t, tt.ss, params[0])

				_, err := fmt.Fprint(w, tt.result)
				require.NoError(t, err)
			}))
			defer svr.Close()

			c := &Client{url: svr.URL}

			got, err := c.SignSequence(tt.ss)
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
	tests := []struct {
		name   string
		hash   common.Hash
		result string
		data   []byte
		err    error
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

				_, err := fmt.Fprint(w, tt.result)
				require.NoError(t, err)
			}))
			defer svr.Close()

			c := &Client{url: svr.URL}

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
