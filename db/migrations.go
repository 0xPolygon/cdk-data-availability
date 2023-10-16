package db

import (
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/gobuffalo/packr/v2"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v4/stdlib"
	migrate "github.com/rubenv/sql-migrate"
)

var packrMigrations = packr.New("migrations", "./migrations")

// RunMigrationsUp runs migrate-up for the given config.
func RunMigrationsUp(pg *pgxpool.Pool) error {
	log.Info("running migrations up")
	return runMigrations(pg, migrate.Up)
}

// RunMigrationsDown runs migrate-down for the given config.
func RunMigrationsDown(pg *pgxpool.Pool) error {
	log.Info("running migrations down")
	return runMigrations(pg, migrate.Down)
}

// runMigrations will execute pending migrations if needed to keep
// the database updated with the latest changes in either direction,
// up or down.
func runMigrations(pg *pgxpool.Pool, direction migrate.MigrationDirection) error {
	_db := stdlib.OpenDB(*pg.Config().ConnConfig)

	var migrations = &migrate.PackrMigrationSource{Box: packrMigrations}
	nMigrations, err := migrate.Exec(_db, "postgres", migrations, direction)
	if err != nil {
		return err
	}

	log.Info("successfully ran ", nMigrations, " migrations")
	return nil
}
