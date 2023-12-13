package db

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/stretchr/testify/require"
)

func Test_DB_StoreOffChainData(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO data_node").
		WithArgs(common.HexToHash("123").Hex(), common.Bytes2Hex([]byte("value"))).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	wdb := sqlx.NewDb(db, "postgres")

	tx, err := wdb.BeginTxx(context.Background(), nil)
	require.NoError(t, err)

	dbPG := New(wdb)

	err = dbPG.StoreOffChainData(context.Background(), []types.OffChainData{{
		Key:   common.HexToHash("123"),
		Value: []byte("value"),
	}}, tx)
	require.NoError(t, err)

	require.NoError(t, tx.Commit())
}
