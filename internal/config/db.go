package config

import "os"

type Dialect string
type MigrationsPath string

const (
	SQLite     Dialect = "sqlite"
	PostgreSQL         = "postgres"
	MySQL              = "mysql"
)

const (
	SQLitePath     MigrationsPath = "./pkg/db/migrations/sqlite"
	PostgreSQLPath                = "./pkg/db/migrations/psql"
	MySQLPath                     = "./pkg/db/migrations/mysql"
)

func GetDialectAndMigrationsPath() (Dialect, MigrationsPath) {
	value := os.Getenv("MTSU_DB")
	switch value {
	case "sqlite":
		return SQLite, SQLitePath
	case "postgres":
		return PostgreSQL, PostgreSQLPath
	case "mysql":
		return MySQL, MySQLPath
	case "mariadb":
		return MySQL, MySQLPath
	default:
		return SQLite, SQLitePath
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
		dialect, _ := GetDialectAndMigrationsPath()
		switch dialect {
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
