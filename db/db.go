package db

import (
	"context"
	"errors"

	"github.com/0xPolygon/cdk-data-availability/offchaindata"
	"github.com/0xPolygon/cdk-validium-node/jsonrpc/types"
	"github.com/0xPolygon/cdk-validium-node/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// DB is the database layer of the data node
type DB struct {
	pg *pgxpool.Pool
}

// New instantiates a DB
func New(pg *pgxpool.Pool) *DB {
	return &DB{
		pg: pg,
	}
}

// BeginStateTransaction begins a DB transaction. The caller is responsible for committing or rolling back the transaction
func (db *DB) BeginStateTransaction(ctx context.Context) (pgx.Tx, error) {
	return db.pg.Begin(ctx)
}

// StoreOffChainData stores and array of key values in the Db
func (db *DB) StoreOffChainData(ctx context.Context, od []offchaindata.OffChainData, dbTx pgx.Tx) error {
	const storeOffChainDataSQL = `
		INSERT INTO data_node.offchain_data (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO NOTHING;
	`
	for _, d := range od {
		if _, err := dbTx.Exec(
			ctx, storeOffChainDataSQL,
			d.Key.Hex(),
			common.Bytes2Hex(d.Value),
		); err != nil {
			return err
		}
	}

	return nil
}

// GetOffChainData returns the value identified by the key
func (db *DB) GetOffChainData(ctx context.Context, key common.Hash, dbTx pgx.Tx) (types.ArgBytes, error) {
	const getOffchainDataSQL = `
		SELECT value
		FROM data_node.offchain_data 
		WHERE key = $1 LIMIT 1;
	`
	var (
		hexValue string
	)

	if err := dbTx.QueryRow(ctx, getOffchainDataSQL, key.Hex()).Scan(&hexValue); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, state.ErrStateNotSynchronized
		}
		return nil, err
	}
	return common.FromHex(hexValue), nil
}

// Exists checks if a key exists in offchain data table
func (db *DB) Exists(ctx context.Context, key common.Hash) bool {
	var keyExists = "SELECT COUNT(*) FROM data_node.offchain_data WHERE key = $1"
	var (
		count uint
	)

	if err := db.pg.QueryRow(ctx, keyExists, key.Hex()).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

// GetLastProcessedBlock returns the latest block successfully processed by the synchronizer
func (db *DB) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	const getLastProcessedBlockSQL = "SELECT max(block) FROM data_node.sync_info;"
	var (
		lastBlock uint64
	)

	if err := db.pg.QueryRow(ctx, getLastProcessedBlockSQL).Scan(&lastBlock); err != nil {
		return 0, err
	}
	return lastBlock, nil
}

// ResetLastProcessedBlock removes all sync_info for blocks greater than `block`
func (db *DB) ResetLastProcessedBlock(ctx context.Context, block uint64) (uint64, error) {
	const resetLastProcessedBlock = "DELETE FROM data_node.sync_info WHERE block > $1"
	var (
		ct  pgconn.CommandTag
		err error
	)
	if ct, err = db.pg.Exec(ctx, resetLastProcessedBlock, block); err != nil {
		return 0, err
	}
	return uint64(ct.RowsAffected()), nil
}

// StoreLastProcessedBlock stores a record of a block processed by the synchronizer
func (db *DB) StoreLastProcessedBlock(ctx context.Context, block uint64, dbTx pgx.Tx) error {
	const storeLastProcessedBlockSQL = `
		INSERT INTO data_node.sync_info (block) 
		VALUES ($1) 
		ON CONFLICT (block) DO UPDATE 
		SET processed = NOW();
	`

	if _, err := dbTx.Exec(ctx, storeLastProcessedBlockSQL, block); err != nil {
		return err
	}
	return nil
}
