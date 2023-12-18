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

			dbPG := New(wdb)

			err = dbPG.StoreOffChainData(context.Background(), tt.od, wdb)
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
	testTable := []struct {
		name      string
		od        []types.OffChainData
		key       common.Hash
		expected  types.ArgBytes
		returnErr error
	}{
		{
			name: "successfully selected value",
			od: []types.OffChainData{{
				Key:   common.HexToHash("key1"),
				Value: []byte("value1"),
			}},
			key:      common.BytesToHash([]byte("key1")),
			expected: []byte("value1"),
		},
		{
			name: "error returned",
			od: []types.OffChainData{{
				Key:   common.HexToHash("key1"),
				Value: []byte("value1"),
			}},
			key:       common.BytesToHash([]byte("key1")),
			returnErr: errors.New("test error"),
		},
		{
			name: "no rows",
			od: []types.OffChainData{{
				Key:   common.HexToHash("key1"),
				Value: []byte("value1"),
			}},
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
			seedOffchainData(t, wdb, mock, tt.od)

			expected := mock.ExpectQuery(`SELECT value FROM data_node\.offchain_data WHERE key = \$1 LIMIT 1`).
				WithArgs(tt.key.Hex())

			if tt.returnErr != nil {
				expected.WillReturnError(tt.returnErr)
			} else {
				expected.WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(common.Bytes2Hex(tt.expected)))
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

func Test_DB_Exist(t *testing.T) {
	testTable := []struct {
		name      string
		od        []types.OffChainData
		key       common.Hash
		count     int
		returnErr error
	}{
		{
			name: "two values found",
			od: []types.OffChainData{{
				Key:   common.HexToHash("key1"),
				Value: []byte("value1"),
			}, {
				Key:   common.HexToHash("key1"),
				Value: []byte("value2"),
			}},
			key:   common.BytesToHash([]byte("key1")),
			count: 2,
		},
		{
			name: "no values found",
			od: []types.OffChainData{{
				Key:   common.HexToHash("key1"),
				Value: []byte("value1"),
			}, {
				Key:   common.HexToHash("key1"),
				Value: []byte("value2"),
			}},
			key:   common.BytesToHash([]byte("undefined")),
			count: 0,
		},
		{
			name: "error returned",
			od: []types.OffChainData{{
				Key:   common.HexToHash("key1"),
				Value: []byte("value1"),
			}},
			key:       common.BytesToHash([]byte("undefined")),
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

			wdb := sqlx.NewDb(db, "postgres")

			// Seed data
			seedOffchainData(t, wdb, mock, tt.od)

			expected := mock.ExpectQuery(`SELECT COUNT\(\*\) FROM data_node\.offchain_data WHERE key = \$1`).
				WithArgs(tt.key.Hex())

			if tt.returnErr != nil {
				expected.WillReturnError(tt.returnErr)
			} else {
				expected.WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(tt.count))
			}

			dbPG := New(wdb)

			actual := dbPG.Exists(context.Background(), tt.key)
			require.NoError(t, err)
			require.Equal(t, tt.count > 0, actual)

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_DB_StoreLastProcessedBlock(t *testing.T) {
	testTable := []struct {
		name      string
		task      string
		block     uint64
		returnErr error
	}{
		{
			name:  "value inserted",
			task:  "task1",
			block: 1,
		},
		{
			name:      "error returned",
			task:      "task1",
			block:     1,
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

			expected := mock.ExpectExec(`INSERT INTO data_node\.sync_tasks \(task, block\) VALUES \(\$1, \$2\) ON CONFLICT \(task\) DO UPDATE SET block = EXCLUDED\.block, processed = NOW\(\)`).
				WithArgs(tt.task, tt.block)
			if tt.returnErr != nil {
				expected.WillReturnError(tt.returnErr)
			} else {
				expected.WillReturnResult(sqlmock.NewResult(1, 1))
			}

			wdb := sqlx.NewDb(db, "postgres")

			dbPG := New(wdb)

			err = dbPG.StoreLastProcessedBlock(context.Background(), tt.task, tt.block, wdb)
			if tt.returnErr != nil {
				require.ErrorIs(t, err, tt.returnErr)
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_DB_GetLastProcessedBlock(t *testing.T) {
	testTable := []struct {
		name      string
		task      string
		block     uint64
		returnErr error
	}{
		{
			name:  "successfully selected block",
			task:  "task1",
			block: 1,
		},
		{
			name:      "error returned",
			task:      "task1",
			block:     1,
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

			mock.ExpectExec(`INSERT INTO data_node\.sync_tasks \(task, block\) VALUES \(\$1, \$2\) ON CONFLICT \(task\) DO UPDATE SET block = EXCLUDED\.block, processed = NOW\(\)`).
				WithArgs(tt.task, tt.block).
				WillReturnResult(sqlmock.NewResult(1, 1))

			expected := mock.ExpectQuery(`SELECT block FROM data_node\.sync_tasks WHERE task = \$1`).
				WithArgs(tt.task)

			if tt.returnErr != nil {
				expected.WillReturnError(tt.returnErr)
			} else {
				expected.WillReturnRows(sqlmock.NewRows([]string{"block"}).AddRow(tt.block))
			}

			wdb := sqlx.NewDb(db, "postgres")

			dbPG := New(wdb)

			err = dbPG.StoreLastProcessedBlock(context.Background(), tt.task, tt.block, wdb)
			require.NoError(t, err)

			actual, err := dbPG.GetLastProcessedBlock(context.Background(), tt.task)
			if tt.returnErr != nil {
				require.ErrorIs(t, err, tt.returnErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.block, actual)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func seedOffchainData(t *testing.T, db *sqlx.DB, mock sqlmock.Sqlmock, od []types.OffChainData) {
	t.Helper()

	mock.ExpectBegin()
	for i, o := range od {
		mock.ExpectExec(`INSERT INTO data_node\.offchain_data \(key, value\) VALUES \(\$1, \$2\) ON CONFLICT \(key\) DO NOTHING`).
			WithArgs(o.Key.Hex(), common.Bytes2Hex(o.Value)).
			WillReturnResult(sqlmock.NewResult(int64(i+1), int64(i+1)))
	}
	mock.ExpectCommit()

	tx, err := db.BeginTxx(context.Background(), nil)
	require.NoError(t, err)

	err = New(db).StoreOffChainData(context.Background(), od, tx)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)
}
