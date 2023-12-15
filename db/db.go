package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jmoiron/sqlx"
)

var (
	// ErrStateNotSynchronized indicates the state database may be empty
	ErrStateNotSynchronized = errors.New("state not synchronized")
)

// IDB defines functions that a DB instance should implement
type IDB interface {
	BeginStateTransaction(ctx context.Context) (IDBTx, error)
	Exists(ctx context.Context, key common.Hash) bool
	GetLastProcessedBlock(ctx context.Context, task string) (uint64, error)
	GetOffChainData(ctx context.Context, key common.Hash, dbTx sqlx.QueryerContext) (types.ArgBytes, error)
	StoreLastProcessedBlock(ctx context.Context, task string, block uint64, dbTx sqlx.ExecerContext) error
	StoreOffChainData(ctx context.Context, od []types.OffChainData, dbTx sqlx.ExecerContext) error
}

// IDBTx is the interface that defines functions a db tx has to implement
type IDBTx interface {
	sqlx.ExecerContext
	sqlx.QueryerContext
	Rollback() error
	Commit() error
}

var _ IDB = (*DB)(nil)

// DB is the database layer of the data node
type DB struct {
	pg *sqlx.DB
}

// New instantiates a DB
func New(pg *sqlx.DB) *DB {
	return &DB{
		pg: pg,
	}
}

// BeginStateTransaction begins a DB transaction. The caller is responsible for committing or rolling back the transaction
func (db *DB) BeginStateTransaction(ctx context.Context) (IDBTx, error) {
	return db.pg.BeginTxx(ctx, nil)
}

// StoreOffChainData stores and array of key values in the Db
func (db *DB) StoreOffChainData(ctx context.Context, od []types.OffChainData, dbTx sqlx.ExecerContext) error {
	const storeOffChainDataSQL = `
		INSERT INTO data_node.offchain_data (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO NOTHING;
	`

	for _, d := range od {
		if _, err := dbTx.ExecContext(
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
func (db *DB) GetOffChainData(ctx context.Context, key common.Hash, dbTx sqlx.QueryerContext) (types.ArgBytes, error) {
	const getOffchainDataSQL = `
		SELECT value
		FROM data_node.offchain_data 
		WHERE key = $1 LIMIT 1;
	`

	var (
		hexValue string
	)

	if err := dbTx.QueryRowxContext(ctx, getOffchainDataSQL, key.Hex()).Scan(&hexValue); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStateNotSynchronized
		}
		return nil, err
	}

	return common.FromHex(hexValue), nil
}

// Exists checks if a key exists in offchain data table
func (db *DB) Exists(ctx context.Context, key common.Hash) bool {
	const keyExists = "SELECT COUNT(*) FROM data_node.offchain_data WHERE key = $1;"

	var (
		count uint
	)

	if err := db.pg.QueryRowContext(ctx, keyExists, key.Hex()).Scan(&count); err != nil {
		return false
	}

	return count > 0
}

// StoreLastProcessedBlock stores a record of a block processed by the synchronizer for named task
func (db *DB) StoreLastProcessedBlock(ctx context.Context, task string, block uint64, dbTx sqlx.ExecerContext) error {
	const storeLastProcessedBlockSQL = `
		INSERT INTO data_node.sync_tasks (task, block) 
		VALUES ($1, $2)
		ON CONFLICT (task) DO UPDATE 
		SET block = EXCLUDED.block, processed = NOW();
	`

	if _, err := dbTx.ExecContext(ctx, storeLastProcessedBlockSQL, task, block); err != nil {
		return err
	}

	return nil
}

// GetLastProcessedBlock returns the latest block successfully processed by the synchronizer for named task
func (db *DB) GetLastProcessedBlock(ctx context.Context, task string) (uint64, error) {
	const getLastProcessedBlockSQL = "SELECT block FROM data_node.sync_tasks WHERE task = $1;"

	var (
		lastBlock uint64
	)

	if err := db.pg.QueryRowContext(ctx, getLastProcessedBlockSQL, task).Scan(&lastBlock); err != nil {
		return 0, err
	}

	return lastBlock, nil
}
