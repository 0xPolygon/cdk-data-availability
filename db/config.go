package db

import (
	"context"
	"fmt"

	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/jmoiron/sqlx"
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
	// DEPRECATED
	EnableLog bool `mapstructure:"EnableLog"`

	// MaxConns is the maximum number of connections in the pool.
	MaxConns int `mapstructure:"MaxConns"`
}

// InitContext initializes DB connection by the given config
func InitContext(ctx context.Context, cfg Config) (*sqlx.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)

	conn, err := sqlx.ConnectContext(ctx, "postgres", psqlInfo)
	if err != nil {
		log.Errorf("Unable to connect to database: %v\n", err)
		return nil, err
	}

	conn.DB.SetMaxIdleConns(cfg.MaxConns)

	if err = conn.PingContext(ctx); err != nil {
		log.Errorf("Unable to ping the database: %v\n", err)
		return nil, err
	}

	return conn, nil
}
