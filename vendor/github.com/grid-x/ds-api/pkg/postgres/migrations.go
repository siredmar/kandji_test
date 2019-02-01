package postgres

import (
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"

	"github.com/grid-x/ds-api/pkg/postgres/migrations"
)

// Migrate migrates the database schema.
func Migrate(db *sqlx.DB) (int, error) {
	var migrations = &migrate.AssetMigrationSource{
		Asset:    migrations.Asset,
		AssetDir: migrations.AssetDir,
		Dir:      "sql",
	}
	return migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
}
