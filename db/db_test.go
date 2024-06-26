package db

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

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

			err = dbPG.StoreLastProcessedBlock(context.Background(), tt.task, tt.block)
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

			err = dbPG.StoreLastProcessedBlock(context.Background(), tt.task, tt.block)
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

func Test_DB_StoreUnresolvedBatchKeys(t *testing.T) {
	testTable := []struct {
		name      string
		bk        []types.BatchKey
		returnErr error
	}{
		{
			name: "no values inserted",
		},
		{
			name: "one value inserted",
			bk: []types.BatchKey{{
				Number: 1,
				Hash:   common.HexToHash("key1"),
			}},
		},
		{
			name: "several values inserted",
			bk: []types.BatchKey{{
				Number: 1,
				Hash:   common.HexToHash("key1"),
			}, {
				Number: 2,
				Hash:   common.HexToHash("key2"),
			}},
		},
		{
			name: "error returned",
			bk: []types.BatchKey{{
				Number: 1,
				Hash:   common.HexToHash("key1"),
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
			for _, o := range tt.bk {
				expected := mock.ExpectExec(`INSERT INTO data_node\.unresolved_batches \(num, hash\) VALUES \(\$1, \$2\) ON CONFLICT \(num, hash\) DO NOTHING`).
					WithArgs(o.Number, o.Hash.Hex())
				if tt.returnErr != nil {
					expected.WillReturnError(tt.returnErr)
				} else {
					expected.WillReturnResult(sqlmock.NewResult(int64(len(tt.bk)), int64(len(tt.bk))))
				}
			}
			if tt.returnErr == nil {
				mock.ExpectCommit()
			} else {
				mock.ExpectRollback()
			}

			wdb := sqlx.NewDb(db, "postgres")

			dbPG := New(wdb)

			err = dbPG.StoreUnresolvedBatchKeys(context.Background(), tt.bk)
			if tt.returnErr != nil {
				require.ErrorIs(t, err, tt.returnErr)
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_DB_GetUnresolvedBatchKeys(t *testing.T) {
	testTable := []struct {
		name      string
		bks       []types.BatchKey
		returnErr error
	}{
		{
			name: "successfully selected data",
			bks: []types.BatchKey{{
				Number: 1,
				Hash:   common.HexToHash("key1"),
			}},
		},
		{
			name: "error returned",
			bks: []types.BatchKey{{
				Number: 1,
				Hash:   common.HexToHash("key1"),
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

			wdb := sqlx.NewDb(db, "postgres")

			// Seed data
			seedUnresolvedBatchKeys(t, wdb, mock, tt.bks)

			expected := mock.ExpectQuery(`SELECT num, hash FROM data_node\.unresolved_batches`)

			if tt.returnErr != nil {
				expected.WillReturnError(tt.returnErr)
			} else {
				for _, bk := range tt.bks {
					expected.WillReturnRows(sqlmock.NewRows([]string{"num", "hash"}).AddRow(bk.Number, bk.Hash.Hex()))
				}
			}

			dbPG := New(wdb)

			data, err := dbPG.GetUnresolvedBatchKeys(context.Background())
			if tt.returnErr != nil {
				require.ErrorIs(t, err, tt.returnErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.bks, data)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_DB_DeleteUnresolvedBatchKeys(t *testing.T) {
	testTable := []struct {
		name      string
		bks       []types.BatchKey
		returnErr error
	}{
		{
			name: "value deleted",
			bks: []types.BatchKey{{
				Number: 1,
				Hash:   common.HexToHash("key1"),
			}},
		},
		{
			name: "error returned",
			bks: []types.BatchKey{{
				Number: 1,
				Hash:   common.HexToHash("key1"),
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
			for _, bk := range tt.bks {
				expected := mock.ExpectExec(`DELETE FROM data_node\.unresolved_batches WHERE num = \$1 AND hash = \$2`).
					WithArgs(bk.Number, bk.Hash.Hex())
				if tt.returnErr != nil {
					expected.WillReturnError(tt.returnErr)
				} else {
					expected.WillReturnResult(sqlmock.NewResult(int64(len(tt.bks)), int64(len(tt.bks))))
				}
			}
			if tt.returnErr != nil {
				mock.ExpectRollback()
			} else {
				mock.ExpectCommit()
			}

			wdb := sqlx.NewDb(db, "postgres")

			dbPG := New(wdb)

			err = dbPG.DeleteUnresolvedBatchKeys(context.Background(), tt.bks)
			if tt.returnErr != nil {
				require.ErrorIs(t, err, tt.returnErr)
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

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
			if tt.returnErr == nil {
				mock.ExpectCommit()
			} else {
				mock.ExpectRollback()
			}

			wdb := sqlx.NewDb(db, "postgres")

			dbPG := New(wdb)

			err = dbPG.StoreOffChainData(context.Background(), tt.od)
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
			key:       common.BytesToHash([]byte("undefined")),
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

			data, err := dbPG.GetOffChainData(context.Background(), tt.key)
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

func Test_DB_ListOffChainData(t *testing.T) {
	testTable := []struct {
		name      string
		od        []types.OffChainData
		keys      []common.Hash
		expected  map[common.Hash]types.ArgBytes
		sql       string
		returnErr error
	}{
		{
			name: "successfully selected one value",
			od: []types.OffChainData{{
				Key:   common.HexToHash("key1"),
				Value: []byte("value1"),
			}},
			keys: []common.Hash{
				common.BytesToHash([]byte("key1")),
			},
			expected: map[common.Hash]types.ArgBytes{
				common.BytesToHash([]byte("key1")): []byte("value1"),
			},
			sql: `SELECT key, value FROM data_node\.offchain_data WHERE key IN \(\$1\)`,
		},
		{
			name: "successfully selected two values",
			od: []types.OffChainData{{
				Key:   common.HexToHash("key1"),
				Value: []byte("value1"),
			}, {
				Key:   common.HexToHash("key2"),
				Value: []byte("value2"),
			}},
			keys: []common.Hash{
				common.BytesToHash([]byte("key1")),
				common.BytesToHash([]byte("key2")),
			},
			expected: map[common.Hash]types.ArgBytes{
				common.BytesToHash([]byte("key1")): []byte("value1"),
				common.BytesToHash([]byte("key2")): []byte("value2"),
			},
			sql: `SELECT key, value FROM data_node\.offchain_data WHERE key IN \(\$1\, \$2\)`,
		},
		{
			name: "error returned",
			od: []types.OffChainData{{
				Key:   common.HexToHash("key1"),
				Value: []byte("value1"),
			}},
			keys: []common.Hash{
				common.BytesToHash([]byte("key1")),
			},
			sql:       `SELECT key, value FROM data_node\.offchain_data WHERE key IN \(\$1\)`,
			returnErr: errors.New("test error"),
		},
		{
			name: "no rows",
			od: []types.OffChainData{{
				Key:   common.HexToHash("key1"),
				Value: []byte("value1"),
			}},
			keys: []common.Hash{
				common.BytesToHash([]byte("undefined")),
			},
			sql:       `SELECT key, value FROM data_node\.offchain_data WHERE key IN \(\$1\)`,
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

			preparedKeys := make([]driver.Value, len(tt.keys))
			for i, key := range tt.keys {
				preparedKeys[i] = key.Hex()
			}

			expected := mock.ExpectQuery(tt.sql).
				WithArgs(preparedKeys...)

			if tt.returnErr != nil {
				expected.WillReturnError(tt.returnErr)
			} else {
				returnData := sqlmock.NewRows([]string{"key", "value"})

				for key, val := range tt.expected {
					returnData = returnData.AddRow(key.Hex(), common.Bytes2Hex(val))
				}

				expected.WillReturnRows(returnData)
			}

			dbPG := New(wdb)

			data, err := dbPG.ListOffChainData(context.Background(), tt.keys)
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

func Test_DB_CountOffchainData(t *testing.T) {
	testTable := []struct {
		name      string
		od        []types.OffChainData
		count     uint64
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
			count: 2,
		},
		{
			name:  "no values found",
			count: 0,
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

			wdb := sqlx.NewDb(db, "postgres")

			// Seed data
			seedOffchainData(t, wdb, mock, tt.od)

			expected := mock.ExpectQuery(`SELECT COUNT\(\*\) FROM data_node\.offchain_data`)

			if tt.returnErr != nil {
				expected.WillReturnError(tt.returnErr)
			} else {
				expected.WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(tt.count))
			}

			dbPG := New(wdb)

			actual, err := dbPG.CountOffchainData(context.Background())
			if tt.returnErr != nil {
				require.ErrorIs(t, err, tt.returnErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.count, actual)
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

	err := New(db).StoreOffChainData(context.Background(), od)
	require.NoError(t, err)
}

func seedUnresolvedBatchKeys(t *testing.T, db *sqlx.DB, mock sqlmock.Sqlmock, bk []types.BatchKey) {
	t.Helper()

	mock.ExpectBegin()
	for i, o := range bk {
		mock.ExpectExec(`INSERT INTO data_node\.unresolved_batches \(num, hash\) VALUES \(\$1, \$2\) ON CONFLICT \(num, hash\) DO NOTHING`).
			WithArgs(o.Number, o.Hash.Hex()).
			WillReturnResult(sqlmock.NewResult(int64(i+1), int64(i+1)))
	}
	mock.ExpectCommit()

	err := New(db).StoreUnresolvedBatchKeys(context.Background(), bk)
	require.NoError(t, err)
}
