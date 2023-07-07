package db

import (
	"context"
	"errors"

	"github.com/0xPolygon/supernets2-data-availability/offchaindata"
	"github.com/0xPolygon/supernets2-node/jsonrpc/types"
	"github.com/0xPolygon/supernets2-node/state"
	"github.com/ethereum/go-ethereum/common"
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
func (d *DB) BeginStateTransaction(ctx context.Context) (pgx.Tx, error) {
	return d.pg.Begin(ctx)
}

// StoreOffChainData stores and array of key valeus in the Db
func (p *DB) StoreOffChainData(ctx context.Context, od []offchaindata.OffChainData, dbTx pgx.Tx) error {
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
func (p *DB) GetOffChainData(ctx context.Context, key common.Hash, dbTx pgx.Tx) (types.ArgBytes, error) {
	const getOffchainDataSQL = `
		SELECT value
		FROM data_node.offchain_data 
		WHERE key = $1 LIMIT 1;
	`
	var (
		valueStr string
	)

	if err := dbTx.QueryRow(ctx, getOffchainDataSQL, key.Hex()).Scan(&valueStr); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, state.ErrStateNotSynchronized
		}
		return nil, err
	}
	return common.Hex2Bytes(valueStr), nil
}

// GetLastProcessedBlock returns the latest block successfully processed by the synchronizer
func (p *DB) GetLastProcessedBlock(ctx context.Context, dbTx pgx.Tx) (uint64, error) {
	const getLastProcessedBlockSQL = "SELECT max(block) from data_node.sync_info;"
	var (
		lastBlock uint64
	)

	if err := dbTx.QueryRow(ctx, getLastProcessedBlockSQL).Scan(&lastBlock); err != nil {
		return 0, err
	}
	return lastBlock, nil
}

// StoreLastProcessedBlock stores a record of a block processed by the synchronizer
func (p *DB) StoreLastProcessedBlock(ctx context.Context, block uint64, dbTx pgx.Tx) error {
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
