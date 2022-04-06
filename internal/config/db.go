package config

import "os"

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

const (
	SQLitePath     MigrationsPath = "./pkg/db/migrations/sqlite"
	PostgreSQLPath                = "./pkg/db/migrations/psql"
	MySQLPath                     = "./pkg/db/migrations/mysql"
)

func GetDialectAndMigrationsPath() DBConfig {
	value := os.Getenv("MTSU_DB")
	switch value {
	case "sqlite":
		return DBConfig{Dialect: SQLite, MigrationsPath: SQLitePath}
	case "postgres":
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
