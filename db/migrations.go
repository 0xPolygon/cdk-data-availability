package db

import (
	"embed"

	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
)

var (
	//go:embed migrations/*.sql
	embedMigrations embed.FS
)

// RunMigrationsUp runs migrate-up for the given config.
func RunMigrationsUp(pg *sqlx.DB) error {
	log.Info("running migrations up")
	return runMigrations(pg, migrate.Up)
}

// runMigrations will execute pending migrations if needed to keep
// the database updated with the latest changes in either direction,
// up or down.
func runMigrations(db *sqlx.DB, direction migrate.MigrationDirection) error {
	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: embedMigrations,
		Root:       "migrations",
	}

	nMigrations, err := migrate.Exec(db.DB, "postgres", migrations, direction)
	if err != nil {
		return err
	}

	log.Info("successfully ran ", nMigrations, " migrations")

	return nil
}
