package sequencer

import (
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

func Test_GetData(t *testing.T) {
	tests := []struct {
		name         string
		batchNum     uint64
		result       string
		expectedData *SeqBatch
		statusCode   int
		err          error
	}{
		{
			name:     "successfully got data",
			batchNum: 10,
			result: fmt.Sprintf(
				`{"result":{"number":"%s","accInputHash":"%s","batchL2Data":"%s"}}`,
				types.ArgUint64(10).Hex(),
				common.BytesToHash([]byte("somedata")),
				types.ArgBytes("l2data").Hex(),
			),
			expectedData: &SeqBatch{
				Number:       10,
				AccInputHash: common.BytesToHash([]byte("somedata")),
				BatchL2Data:  []byte("l2data"),
			},
		},
		{
			name:     "error returned by server",
			batchNum: 10,
			result:   `{"error":{"code":123,"message":"test error"}}`,
			err:      errors.New("123 - test error"),
		},
		{
			name:     "invalid data returned by server",
			batchNum: 10,
			result:   `{"result":"invalid-signature"}`,
			err:      errors.New("json: cannot unmarshal string into Go value of type sequencer.SeqBatch"),
		},
		{
			name:       "unsuccessful status code returned by server",
			batchNum:   10,
			statusCode: http.StatusInternalServerError,
			err:        errors.New("invalid status code, expected: 200, found: 500"),
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var res rpc.Request
				require.NoError(t, json.NewDecoder(r.Body).Decode(&res))
				require.Equal(t, "zkevm_getBatchByNumber", res.Method)

				var params []interface{}
				require.NoError(t, json.Unmarshal(res.Params, &params))
				require.Equal(t, float64(tt.batchNum), params[0])
				require.True(t, params[1].(bool))

				if tt.statusCode > 0 {
					w.WriteHeader(tt.statusCode)
				}

				_, err := fmt.Fprint(w, tt.result)
				require.NoError(t, err)
			}))
			defer svr.Close()

			got, err := GetData(svr.URL, tt.batchNum)
			if tt.err != nil {
				require.Error(t, err)
				require.EqualError(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedData, got)
			}
		})
	}
}
