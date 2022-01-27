package db

import (
	"database/sql"
	"github.com/Mangatsu/server/internal/config"
	"github.com/pressly/goose/v3"
	log "github.com/sirupsen/logrus"
	"os"
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
		log.Error("Invalid DB driver", "driver", "sqlite3", err)
		os.Exit(1)
	}

	err = goose.Run("up", db(), "./pkg/db/migrations")
	if err != nil {
		log.Error("Failed to apply new migrations", err)
		os.Exit(1)
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
