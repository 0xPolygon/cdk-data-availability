package db

import (
	"context"
	"fmt"

	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Config provide fields to configure the pool
type Config struct {
	// Database name
	Name string `mapstructure:"Name"`

	// Database User name
	User string `mapstructure:"User"`

	// Database Password of the user
	Password string `mapstructure:"Password"`

	// Host address of database
	Host string `mapstructure:"Host"`

	// Port Number of database
	Port string `mapstructure:"Port"`

	// EnableLog
	EnableLog bool `mapstructure:"EnableLog"`

	// MaxConns is the maximum number of connections in the pool.
	MaxConns int `mapstructure:"MaxConns"`
}

// NewSQLDB creates a new SQL DB
func NewSQLDB(cfg Config) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(fmt.Sprintf("postgres://%s:%s@%s:%s/%s?pool_max_conns=%d", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.MaxConns))
	if err != nil {
		log.Errorf("Unable to parse DB config: %v\n", err)
		return nil, err
	}
	if cfg.EnableLog {
		config.ConnConfig.Logger = logger{}
	}
	conn, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		log.Errorf("Unable to connect to database: %v\n", err)
		return nil, err
	}
	return conn, nil
}
