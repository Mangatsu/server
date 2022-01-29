package db

import (
	"database/sql"
	"github.com/Mangatsu/server/internal/config"
	"github.com/pressly/goose/v3"
	log "github.com/sirupsen/logrus"
)

var database *sql.DB

// Initdb initializes the database
func Initdb() {
	var dbErr error
	database, dbErr = sql.Open("sqlite3", config.BuildDataPath("mangatsu.sqlite"))
	if dbErr != nil {
		log.Fatal(dbErr)
	}
}

func db() *sql.DB {
	return database
}

// EnsureLatestVersion ensures that the database is at the latest version by running all migrations.
func EnsureLatestVersion() {
	err := goose.SetDialect("sqlite3")
	if err != nil {
		log.Fatal("Invalid DB driver", "driver", "sqlite3", err)
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
