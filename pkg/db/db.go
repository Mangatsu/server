package db

import (
	"database/sql"
	"embed"
	"github.com/Mangatsu/server/internal/config"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	log "github.com/sirupsen/logrus"
)

var database *sql.DB

//go:embed migrations
var embedMigrations embed.FS

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
	// For embedding the migrations in the binary.
	goose.SetBaseFS(embedMigrations)

	dbConfig := config.GetDialectAndMigrationsPath()
	if err := goose.SetDialect(string(dbConfig.Dialect)); err != nil {
		log.Fatal("Invalid DB dialect: ", dbConfig.Dialect, ".", err)
	}

	if err := goose.Up(db(), string(dbConfig.MigrationsPath)); err != nil {
		log.Fatal("Failed to apply new migrations", err)
	}
}

func rollbackTx(tx *sql.Tx) {
	err := tx.Rollback()
	if err != nil {
		log.Debug("Failed to rollback transaction", err)
	}
}
