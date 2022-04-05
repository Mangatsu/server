package db

import (
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	log "github.com/sirupsen/logrus"
)

// Database is a wrapper around the database connection handle
// and it stores its dialect
type Database struct {
	Dialect string
	MigrationsPath string
	Handle *sql.DB
	DialectWrapper goqu.DialectWrapper
}

var database *Database

// Initdb initializes the database
func Initdb(dialect, connString string) *Database {
	handle, err := sql.Open(dialect, connString)
	if err != nil {
		log.Fatal(err)
	}

	migrationsPath := GetMigrationsPath(dialect)
	dialectWrapper := goqu.Dialect(dialect)

	return &Database{dialect, migrationsPath, handle, dialectWrapper}
}

// QB returns a query builder for the database
func (db *Database) QB() *goqu.Database {
	return db.DialectWrapper.DB(db.Handle)
}

// GetMigrationsPath returns path for the given dialect,
// or an empty string if it's not known
func GetMigrationsPath(dialect string) string {
	// TODO: change hard-coded case strings
	switch dialect {
	case "sqlite3":
		return "./pkg/db/migrations/sqlite"
	case "postgres":
		return "./pkg/db/migrations/psql"
	default:
		return ""
	}
}

func db() *sql.DB {
	// FIXME: it's still here only not to break current code
	return database.Handle
}

// EnsureLatestVersion ensures that the database is at the latest version by running all migrations.
func EnsureLatestVersion() {
	err := goose.SetDialect(database.Dialect)
	if err != nil {
		log.Fatal("Invalid DB driver", "driver", database.Dialect, err)
	}

	err = goose.Run("up", database.Handle, database.MigrationsPath)
	if err != nil {
		log.Fatal("Failed to apply new migrations", err)
	}
}

func rollbackTx(tx *sql.Tx) {
	err := tx.Rollback()
	if err != nil {
		log.Debug("Failed to rollback transaction", err)
	}
}

func Clamp(value, min, max int64) int64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
