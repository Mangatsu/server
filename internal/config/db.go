package config

import "os"

const (
	SQLite     string = "sqlite"
	PostgreSQL        = "postgres"
	MySQL             = "mysql"
)

func GetDBDriver() string {
	value := os.Getenv("MTSU_DB")
	switch value {
	case "sqlite":
		return SQLite
	case "postgres":
		return PostgreSQL
	case "mysql":
		return MySQL
	case "mariadb":
		return MySQL
	default:
		return SQLite
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
		driver := GetDBDriver()
		switch driver {
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
