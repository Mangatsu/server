package db

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/log"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

var database *sql.DB

//go:embed migrations
var embedMigrations embed.FS

// InitDB initializes the database
func InitDB() {
	var dbErr error
	database, dbErr = sql.Open("sqlite3", config.BuildDataPath(config.Options.DB.Name+".sqlite"))
	if dbErr != nil {
		log.Z.Fatal(dbErr.Error())
	}
}

func db() *sql.DB {
	return database
}

// EnsureLatestVersion ensures that the database is at the latest version by running all migrations.
func EnsureLatestVersion() {
	if !config.Options.DB.Migrations {
		log.Z.Warn("database migrations are disabled.")
		return
	}

	// For embedding the migrations in the binary.
	goose.SetBaseFS(embedMigrations)

	err := goose.SetDialect("sqlite3")
	if err != nil {
		log.Z.Fatal("failed setting DB dialect", zap.String("err", err.Error()))
	}

	err = goose.Up(db(), "migrations")
	fmt.Println("")
	if err != nil {
		log.Z.Fatal("failed to apply new migrations", zap.String("err", err.Error()))
	}
}

func rollbackTx(tx *sql.Tx) {
	err := tx.Rollback()
	if err != nil {
		log.Z.Debug("failed to rollback transaction", zap.String("err", err.Error()))
	}
}
