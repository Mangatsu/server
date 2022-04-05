package db

import (
	"database/sql"
	"github.com/Mangatsu/server/internal/config"
	_ "github.com/mattn/go-sqlite3"
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
	dialect, migrationsPath := config.GetDialectAndMigrationsPath()
	if err := goose.SetDialect(string(dialect)); err != nil {
		log.Fatal("Invalid DB dialect: ", dialect, ".", err)
	}

	if err := goose.Run("up", db(), string(migrationsPath)); err != nil {
		log.Fatal("Failed to apply new migrations", err)
	}
}

func rollbackTx(tx *sql.Tx) {
	err := tx.Rollback()
	if err != nil {
		log.Debug("Failed to rollback transaction", err)
	}
}
