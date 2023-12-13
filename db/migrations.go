package db

import (
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/gobuffalo/packr/v2"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
)

var packrMigrations = packr.New("migrations", "./migrations")

// RunMigrationsUp runs migrate-up for the given config.
func RunMigrationsUp(pg *sqlx.DB) error {
	log.Info("running migrations up")
	return runMigrations(pg, migrate.Up)
}

// runMigrations will execute pending migrations if needed to keep
// the database updated with the latest changes in either direction,
// up or down.
func runMigrations(db *sqlx.DB, direction migrate.MigrationDirection) error {
	var migrations = &migrate.PackrMigrationSource{Box: packrMigrations}
	nMigrations, err := migrate.Exec(db.DB, "postgres", migrations, direction)
	if err != nil {
		return err
	}

	log.Info("successfully ran ", nMigrations, " migrations")
	return nil
}
