package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Mangatsu/server/pkg/log"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

type ImageFormat string

const (
	WebP ImageFormat = "webp"
	AVIF             = "avif"
)

type CacheOptions struct {
	WebServer bool
	TTL       time.Duration
	Size      uint64
}

type GalleryOptions struct {
	ThumbnailFormat       ImageFormat
	FuzzySearchSimilarity float64
}

type OptionsModel struct {
	Environment    log.Environment
	Domain         string
	Hostname       string
	Port           string
	StrictACAO     bool
	Registrations  bool
	Visibility     Visibility
	DB             DBOptions
	Cache          CacheOptions
	GalleryOptions GalleryOptions
}

type CredentialsModel struct {
	JWTSecret  string
	Passphrase string
}

var AppEnvironment log.Environment
var LogLevel zapcore.Level

// Options stores the global configuration for the server
var Options *OptionsModel

// Credentials stores the JWT secret, and optionally a passphrase and credentials for the db
var Credentials *CredentialsModel // TODO: Encrypt in memory?

// LoadEnv loads the environment variables.
func LoadEnv() {
	var err = godotenv.Load()
	if err != nil {
		fmt.Println("No .env file or environmentals found.")
	}

	loadEnvironment()
	loadLogLevel()
}

// SetEnv sets the environment variables into Options and Credentials
func SetEnv() {
	Options = &OptionsModel{
		Domain:        domain(),
		Hostname:      hostname(),
		Port:          port(),
		StrictACAO:    acao(),
		Registrations: registrationsEnabled(),
		Visibility:    currentVisibility(),
		DB: DBOptions{
			Name:       dbName(),
			Migrations: dbMigrationsEnabled(),
		},
		Cache: CacheOptions{
			WebServer: cacheServerEnabled(),
			TTL:       cacheTTL(),
			Size:      cacheSize(),
		},
		GalleryOptions: GalleryOptions{
			ThumbnailFormat:       thumbnailFormat(),
			FuzzySearchSimilarity: fuzzySearchSimilarity(),
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

func loadEnvironment() {
	value := os.Getenv("MTSU_ENV")
	if value == "production" {
		AppEnvironment = log.Production
		return
	}
	AppEnvironment = log.Development
}

func loadLogLevel() {
	value := os.Getenv("MTSU_LOG_LEVEL")
	switch value {
	case "debug":
		LogLevel = zap.DebugLevel
	case "warn":
		LogLevel = zap.WarnLevel
	case "error":
		LogLevel = zap.ErrorLevel
	default:
		LogLevel = zap.InfoLevel
	}
}

func domain() string {
	return os.Getenv("MTSU_DOMAIN")
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

func acao() bool {
	return os.Getenv("MTSU_STRICT_ACAO") == "true"
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
			log.Z.Warn("MTSU_RESTRICTED_PASSPHRASE is not set. Defaulting to 's3cr3t'.")
		}
		return "s3cr3t"
	}
	return value
}

func jwtSecret() string {
	value := os.Getenv("MTSU_JWT_SECRET")
	if value == "" {
		log.Z.Error("MTSU_JWT_SECRET is not set. An unsecure secret will be used instead. DO NOT USE IN PRODUCTION.")
		return "9Wag7sMvKl3aF6K5lwIg6TI42ia2f6BstZAVrdJIq8Mp38lnl7UzQMC1qjKyZCBzHFGbbqsA0gKcHqDuyXQAhWoJ0lcx4K5q"
	}
	return value
}

func cacheTTL() time.Duration {
	defaultDuration := time.Hour * 336
	minDuration := time.Minute * 15

	value := os.Getenv("MTSU_CACHE_TTL")
	if value == "" {
		return defaultDuration
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		log.Z.Error(value + " is not a valid TTL for MTSU_CACHE_TTL. Defaulting to 336h (14 days).")
		return defaultDuration
	}

	if duration < minDuration {
		log.Z.Info("Minimum TTL is 15 minutes. Defaulting to 15 minutes.")
		return minDuration
	}

	return duration
}

func cacheSize() uint64 {
	defaultSize := uint64(20000)
	minSize := uint64(100)

	value := os.Getenv("MTSU_CACHE_SIZE")
	if value == "" {
		return defaultSize
	}

	size, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		log.Z.Error(value + " is not a valid TTL for MTSU_CACHE_SIZE. Defaulting to 20 000 MiB.")
		return defaultSize
	}

	if size < minSize {
		log.Z.Warn("Minimum TTL is 100 MiB. Defaulting to 100 MiB.")
		return minSize
	}

	return size
}

func thumbnailFormat() ImageFormat {
	// TODO: Add support for AVIF
	//value := os.Getenv("MTSU_THUMBNAIL_FORMAT")
	//if value == "avif" {
	//	return AVIF
	//}
	return WebP
}

func fuzzySearchSimilarity() float64 {
	value := os.Getenv("MTSU_FUZZY_SEARCH_SIMILARITY")
	if value == "" {
		return 0.7
	}

	similarity, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Z.Error(value + " is not a valid similarity value for MTSU_FUZZY_SEARCH_SIMILARITY. Defaulting to 0.7.")
		return 0.7
	}
	if similarity < 0.1 {
		return 0.1
	}
	if similarity > 1 {
		return 1
	}
	return similarity
}
