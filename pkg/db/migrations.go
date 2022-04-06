package db

import (
	"embed"
	"github.com/Mangatsu/server/internal/config"
	"github.com/pressly/goose/v3"
	log "github.com/sirupsen/logrus"
)

type MigrationsPath string

// Paths specified here point to the embedded migrations-directory in the binary. See var embedMigrations.
const (
	SQLitePath     MigrationsPath = "migrations/sqlite"
	PostgreSQLPath                = "migrations/psql"
	MySQLPath                     = "migrations/mysql"
)

//go:embed migrations
var embedMigrations embed.FS

// getMigrationsPath returns path for the given dialect, or an empty string if it's not known
func getMigrationsPath(dialect config.Dialect) MigrationsPath {
	switch dialect {
	case config.SQLite:
		return SQLitePath
	case config.PostgreSQL:
		return PostgreSQLPath
	case config.MySQL:
		return MySQLPath
	default:
		return ""
	}
}

// EnsureLatestVersion ensures that the database is at the latest version by running all migrations.
func EnsureLatestVersion() {
	if database == nil || database.MigrationsPath == "" {
		log.Fatal("Database is not initialized.")
	}

	if !config.Options.DB.Migrations {
		log.Warning("Database migrations are disabled.")
		return
	}

	// For embedding the migrations in the binary.
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect(string(database.Dialect)); err != nil {
		log.Fatal("Invalid DB dialect: ", database.Dialect, ".", err)
	}

	if err := goose.Up(database.Handle, string(database.MigrationsPath)); err != nil {
		log.Fatal("Failed to apply new migrations", err)
	}
}
