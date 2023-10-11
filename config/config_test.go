package config

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-data-availability/config/types"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func Test_Defaults(t *testing.T) {
	tcs := []struct {
		path          string
		expectedValue interface{}
	}{
		{
			path:          "L1.WsURL",
			expectedValue: "ws://127.0.0.1:8546",
		},
		{
			path:          "L1.RpcURL",
			expectedValue: "http://127.0.0.1:8545",
		},
		{
			path:          "L1.CDKValidiumAddress",
			expectedValue: "0x0DCd1Bf9A1b36cE34237eEaFef220932846BCD82",
		},
		{
			path:          "L1.Timeout",
			expectedValue: types.NewDuration(1 * time.Minute),
		},
		{
			path:          "L1.RetryPeriod",
			expectedValue: types.NewDuration(5 * time.Second),
		},
		{
			path:          "L1.BlockBatchSize",
			expectedValue: uint(64),
		},
		// TODO: more default checks
	}

	ctx := cli.NewContext(cli.NewApp(), nil, nil)
	cfg, err := Load(ctx)
	require.NoError(t, err)

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			actual := getValueFromStruct(tc.path, cfg)
			require.Equal(t, tc.expectedValue, actual)
		})
	}
}

func getValueFromStruct(path string, object interface{}) interface{} {
	keySlice := strings.Split(path, ".")
	v := reflect.ValueOf(object)

	for _, key := range keySlice {
		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		v = v.FieldByName(key)
	}
	return v.Interface()
}
