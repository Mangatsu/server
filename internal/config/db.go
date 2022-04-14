package config

import (
	"os"
)

type DBOptions struct {
	Name       string
	Migrations bool
}

func dbMigrationsEnabled() bool {
	// Disabled only when explicitly set to false.
	return os.Getenv("MTSU_DB_MIGRATIONS") != "false"
}

func dbName() string {
	value := os.Getenv("MTSU_DB_NAME")
	if value == "" {
		return "mangatsu"
	}
	return value
}
