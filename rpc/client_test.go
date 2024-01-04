package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_JSONRPCCallWithContext(t *testing.T) {
	tests := []struct {
		name       string
		result     string
		resp       Response
		statusCode int
		err        error
	}{
		{
			name:   "successfully executed request",
			result: `{"result":"test"}`,
			resp: Response{
				Result: json.RawMessage(`"test"`),
			},
		},
		{
			name:       "unsuccessful status code returned by server",
			statusCode: http.StatusUnauthorized,
			err:        errors.New("invalid status code, expected: 200, found: 401"),
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var res Request
				require.NoError(t, json.NewDecoder(r.Body).Decode(&res))
				require.Equal(t, "test", res.Method)

				if tt.statusCode > 0 {
					w.WriteHeader(tt.statusCode)
				}

				_, err := fmt.Fprint(w, tt.result)
				require.NoError(t, err)
			}))
			defer svr.Close()

			got, err := JSONRPCCallWithContext(context.Background(), svr.URL, "test")
			if tt.err != nil {
				require.Error(t, err)
				require.EqualError(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.resp, got)
			}
		})
	}
}
