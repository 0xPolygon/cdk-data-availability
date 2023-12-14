package db

import (
	"context"
	"errors"
	"testing"

	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func Test_DB_StoreOffChainData(t *testing.T) {
	testTable := []struct {
		name      string
		od        []types.OffChainData
		returnErr error
	}{
		{
			name: "no values inserted",
		},
		{
			name: "one value inserted",
			od: []types.OffChainData{{
				Key:   common.HexToHash("key1"),
				Value: []byte("value1"),
			}},
		},
		{
			name: "several values inserted",
			od: []types.OffChainData{{
				Key:   common.HexToHash("key1"),
				Value: []byte("value1"),
			}, {
				Key:   common.HexToHash("key2"),
				Value: []byte("value2"),
			}},
		},
		{
			name: "error returned",
			od: []types.OffChainData{{
				Key:   common.HexToHash("key1"),
				Value: []byte("value1"),
			}},
			returnErr: errors.New("test error"),
		},
	}

	for _, tt := range testTable {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)

			defer db.Close()

			mock.ExpectBegin()
			for _, o := range tt.od {
				expected := mock.ExpectExec(`INSERT INTO data_node\.offchain_data \(key, value\) VALUES \(\$1, \$2\) ON CONFLICT \(key\) DO NOTHING`).
					WithArgs(o.Key.Hex(), common.Bytes2Hex(o.Value))
				if tt.returnErr != nil {
					expected.WillReturnError(tt.returnErr)
				} else {
					expected.WillReturnResult(sqlmock.NewResult(int64(len(tt.od)), int64(len(tt.od))))
				}
			}

			wdb := sqlx.NewDb(db, "postgres")

			tx, err := wdb.BeginTxx(context.Background(), nil)
			require.NoError(t, err)

			dbPG := New(wdb)

			err = dbPG.StoreOffChainData(context.Background(), tt.od, tx)
			if tt.returnErr != nil {
				require.ErrorIs(t, err, tt.returnErr)
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
