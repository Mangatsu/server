package db

import (
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	log "github.com/sirupsen/logrus"
)

type Database struct {
	Dialect string
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

	dialectWrapper := goqu.Dialect(dialect)

	return &Database{dialect, handle, dialectWrapper}
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

	err = goose.Run("up", db(), "./pkg/db/migrations")
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
