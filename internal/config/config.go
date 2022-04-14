package config

import (
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"os"
)

type Layout string

const (
	Freeform   Layout = "freeform"
	Structured        = "structured"
)

type Visibility string

const (
	Private    Visibility = "private"
	Restricted            = "restricted"
	Public                = "public"
)

type OptionsModel struct {
	Hostname      string
	Port          string
	CacheServer   bool
	Registrations bool
	Visibility    Visibility
	DB            DBOptions
}

type CredentialsModel struct {
	JWTSecret  string
	Passphrase string
}

// Options stores the global configuration for the server
var Options *OptionsModel

// Credentials stores the JWT secret, and optionally a passphrase and credentials for the db
var Credentials *CredentialsModel // TODO: Encrypt in memory?

// LoadEnv loads the environment variables into Options and Credentials
func LoadEnv() {
	var err = godotenv.Load()
	if err != nil {
		log.Debug("No .env file found")
	}

	Options = &OptionsModel{
		Hostname:      hostname(),
		Port:          port(),
		CacheServer:   cacheServerEnabled(),
		Registrations: registrationsEnabled(),
		Visibility:    currentVisibility(),
		DB: DBOptions{
			Name:       dbName(),
			Migrations: dbMigrationsEnabled(),
		},
	}

	Credentials = &CredentialsModel{
		JWTSecret:  jwtSecret(),
		Passphrase: restrictedPassphrase(),
	}
}

func GetInitialAdmin() (string, string) {
	username := os.Getenv("MTSU_INITIAL_ADMIN_NAME")
	password := os.Getenv("MTSU_INITIAL_ADMIN_PW")
	if username == "" {
		username = "admin"
	}
	if password == "" {
		password = "admin321"
	}
	return username, password
}

func hostname() string {
	value := os.Getenv("MTSU_HOSTNAME")
	if value == "" {
		return "localhost"
	}
	return value
}

func port() string {
	value := os.Getenv("MTSU_PORT")
	if value == "" {
		return "5050"
	}
	return value
}

func cacheServerEnabled() bool {
	value := os.Getenv("MTSU_DISABLE_CACHE_SERVER")
	if value == "true" {
		return false
	}
	return true
}

func registrationsEnabled() bool {
	value := os.Getenv("MTSU_REGISTRATIONS")
	if value == "true" {
		return true
	}
	return false
}

func currentVisibility() Visibility {
	value := os.Getenv("MTSU_VISIBILITY")
	switch value {
	case "public":
		return Public
	case "restricted":
		return Restricted
	default:
		return Private
	}
}

func restrictedPassphrase() string {
	value := os.Getenv("MTSU_RESTRICTED_PASSPHRASE")
	if value == "" {
		if currentVisibility() == Restricted {
			log.Error("MTSU_RESTRICTED_PASSPHRASE is not set. Defaulting to 's3cr3t'.")
		}
		return "s3cr3t"
	}
	return value
}

func jwtSecret() string {
	value := os.Getenv("MTSU_JWT_SECRET")
	if value == "" {
		log.Error("MTSU_JWT_SECRET is not set. An unsecure secret will be used instead. DO NOT USE IN PRODUCTION.")
		return "iugnrg8o9347ghjmloi2jhbaw8723hjdbjnwq"
	}
	return value
}
