package config

import (
	"os"
	"strings"
)

type Dialect string
type MigrationsPath string

type DBConfig struct {
	Dialect        Dialect
	MigrationsPath MigrationsPath
}

const (
	SQLite     Dialect = "sqlite"
	PostgreSQL         = "postgres"
	MySQL              = "mysql"
)

// Paths specified here point to the embedded migrations-directory in the binary.
const (
	SQLitePath     MigrationsPath = "migrations/sqlite"
	PostgreSQLPath                = "migrations/psql"
	MySQLPath                     = "migrations/mysql"
)

func GetDialectAndMigrationsPath() DBConfig {
	value := strings.ToLower(os.Getenv("MTSU_DB"))
	switch value {
	case "sqlite", "sqlite3":
		return DBConfig{Dialect: SQLite, MigrationsPath: SQLitePath}
	case "postgres", "postgresql", "psql":
		return DBConfig{Dialect: PostgreSQL, MigrationsPath: PostgreSQLPath}
	case "mysql", "mariadb":
		return DBConfig{Dialect: MySQL, MigrationsPath: MySQLPath}
	default:
		return DBConfig{Dialect: SQLite, MigrationsPath: SQLitePath}
	}
}

func GetDBName() string {
	value := os.Getenv("MTSU_DB_NAME")
	if value == "" {
		return "mtsu"
	}
	return value
}

func GetDBUser() string {
	return os.Getenv("MTSU_DB_USER")
}

func GetDBPassword() string {
	return os.Getenv("MTSU_DB_PASSWORD")
}

func GetDBHost() string {
	value := os.Getenv("MTSU_DB_HOST")
	if value == "" {
		return "localhost"
	}
	return value
}

func GetDBPort() string {
	value := os.Getenv("MTSU_DB_PORT")
	if value == "" {
		dbConfig := GetDialectAndMigrationsPath()
		switch dbConfig.Dialect {
		case PostgreSQL:
			return "5432"
		case MySQL:
			return "3306"
		default:
			return ""
		}
	}
	return value
}
