package db

import (
	"context"
	"errors"
	"testing"

	"github.com/0xPolygon/cdk-data-availability/rpc"

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

func Test_DB_GetOffChainData(t *testing.T) {
	od := []types.OffChainData{{
		Key:   common.HexToHash("key1"),
		Value: []byte("value1"),
	}}

	seedOffchainData := func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO data_node\.offchain_data \(key, value\) VALUES \(\$1, \$2\) ON CONFLICT \(key\) DO NOTHING`).
			WithArgs(od[0].Key.Hex(), common.Bytes2Hex(od[0].Value)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.BeginTxx(context.Background(), nil)
		require.NoError(t, err)

		err = New(db).StoreOffChainData(context.Background(), od, tx)
		require.NoError(t, err)

		err = tx.Commit()
		require.NoError(t, err)
	}

	testTable := []struct {
		name      string
		key       common.Hash
		expected  rpc.ArgBytes
		returnErr error
	}{
		{
			name:     "successfully selected value",
			key:      od[0].Key,
			expected: od[0].Value,
		},
		{
			name:      "error returned",
			key:       od[0].Key,
			returnErr: errors.New("test error"),
		},
		{
			name:      "no rows",
			key:       common.BytesToHash([]byte("underfined")),
			returnErr: ErrStateNotSynchronized,
		},
	}

	for _, tt := range testTable {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)

			defer db.Close()

			wdb := sqlx.NewDb(db, "postgres")

			// Seed data
			seedOffchainData(wdb, mock)

			expected := mock.ExpectQuery(`SELECT value FROM data_node\.offchain_data WHERE key = \$1 LIMIT 1`).
				WithArgs(tt.key.Hex())

			if tt.returnErr != nil {
				expected = expected.WillReturnError(tt.returnErr)
			} else {
				expected = expected.WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(common.Bytes2Hex(tt.expected)))
			}

			dbPG := New(wdb)

			data, err := dbPG.GetOffChainData(context.Background(), tt.key, wdb)
			if tt.returnErr != nil {
				require.ErrorIs(t, err, tt.returnErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, data)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
