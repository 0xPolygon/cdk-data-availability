package client

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var res rpc.Request
				require.NoError(t, json.NewDecoder(r.Body).Decode(&res))

				var params [][]types.SignedSequence
				require.NoError(t, json.Unmarshal(res.Params, &params))
				require.Equal(t, tt.ss, params[0][0])

				_, err := fmt.Fprint(w, tt.result)
				require.NoError(t, err)
			}))
			defer svr.Close()

			c := &Client{url: svr.URL}

			got, err := c.SignSequence(tt.ss)
			if tt.err != nil {
				require.EqualError(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.signature, got)
			}
		})
	}
}
