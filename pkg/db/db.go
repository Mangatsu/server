package db

import (
	"database/sql"
	"github.com/Mangatsu/server/internal/config"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

// Database is a wrapper around the database connection handle, and it stores its dialect
type Database struct {
	Dialect        config.Dialect
	MigrationsPath MigrationsPath
	Handle         *sql.DB
	DialectWrapper goqu.DialectWrapper
}

var database *Database

// InitDB initializes the database
func InitDB() {
	dialect := config.Options.DB.Dialect
	connString := buildConnString(dialect)

	handle, err := sql.Open(string(dialect), connString)
	if err != nil {
		log.Fatal(err)
	}

	migrationsPath := getMigrationsPath(dialect)
	dialectWrapper := goqu.Dialect(string(dialect))

	database = &Database{dialect, migrationsPath, handle, dialectWrapper}
}

// QB returns a query builder for the database
func (db *Database) QB() *goqu.Database {
	return db.DialectWrapper.DB(db.Handle)
}

// buildConnString builds the connection string for specified dialect.
// If no dialect specified is not valid, an empty string is returned.
func buildConnString(dialect config.Dialect) string {
	// TODO: Add support for other dialects
	switch dialect {
	case config.SQLite:
		return config.BuildDataPath(config.Options.DB.Name + ".sqlite")
	case config.PostgreSQL:
		return ""
	case config.MySQL:
		return ""
	default:
		return ""
	}
}

func db() *sql.DB {
	// FIXME: it's still here only not to break current code
	return database.Handle
}

func rollbackTx(tx *sql.Tx) {
	err := tx.Rollback()
	if err != nil {
		log.Debug("Failed to rollback transaction", err)
	}
}
