package config

import (
	"os"
	"strings"
)

type Dialect string

const (
	SQLite     Dialect = "sqlite3"
	PostgreSQL         = "postgres"
	MySQL              = "mysql"
)

type DBOptions struct {
	Dialect    Dialect
	Host       string
	Port       string
	Name       string
	Migrations bool
}

type DBCredentials struct {
	User     string
	Password string
}

func dbMigrationsEnabled() bool {
	// Disabled only when explicitly set to false.
	return os.Getenv("MTSU_DB_MIGRATIONS") != "false"
}

func dbDialect() Dialect {
	value := strings.ToLower(os.Getenv("MTSU_DB"))
	switch value {
	case "sqlite", "sqlite3":
		return SQLite
	case "postgres", "postgresql", "psql":
		return PostgreSQL
	case "mysql", "mariadb":
		return MySQL
	default:
		return SQLite
	}
}

func dbName() string {
	value := os.Getenv("MTSU_DB_NAME")
	if value == "" {
		return "mangatsu"
	}
	return value
}

func dbUser() string {
	return os.Getenv("MTSU_DB_USER")
}

func dbPassword() string {
	return os.Getenv("MTSU_DB_PASSWORD")
}

func dbHost() string {
	value := os.Getenv("MTSU_DB_HOST")
	if value == "" {
		return "localhost"
	}
	return value
}

func dbPort() string {
	value := os.Getenv("MTSU_DB_PORT")
	if value == "" {
		switch dbDialect() {
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
