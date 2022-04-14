package config

import (
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
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

type CacheOptions struct {
	WebServer bool
	TTL       time.Duration
}

type OptionsModel struct {
	Hostname      string
	Port          string
	Registrations bool
	Visibility    Visibility
	DB            DBOptions
	Cache         CacheOptions
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
		Registrations: registrationsEnabled(),
		Visibility:    currentVisibility(),
		DB: DBOptions{
			Name:       dbName(),
			Migrations: dbMigrationsEnabled(),
		},
		Cache: CacheOptions{
			WebServer: cacheServerEnabled(),
			TTL:       cacheTTL(),
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

func cacheTTL() time.Duration {
	value := os.Getenv("MTSU_CACHE_TTL")
	if value == "" {
		value = "336h"
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		log.Warningf("%s is not a valid TTL for MTSU_CACHE_TTL. Defaulting to 336h (14 days)", value)
		duration, _ = time.ParseDuration("336h")
	}

	if duration < time.Minute*15 {
		log.Warning("Minimum TTL is 15 minutes. Defaulting to 15m")
		duration, _ = time.ParseDuration("10s")
	}

	return duration
}
