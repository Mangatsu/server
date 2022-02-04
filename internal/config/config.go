package config

import (
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Layout string

const (
	Freeform   Layout = "freeform"
	Structured        = "structured"
)

const (
	Private    string = "private"
	Restricted        = "restricted"
	Public            = "public"
)

func LoadEnv() {
	var err = godotenv.Load()
	if err != nil {
		log.Debug("No .env file found")
	}
}

type Library struct {
	ID     int32
	Path   string
	Layout string
}

var libraryOptionsR = regexp.MustCompile(`^(freeform|structured)(\d+)$`)

func ParseBasePaths() []Library {
	basePaths := os.Getenv("MTSU_BASE_PATHS")
	if basePaths == "" {
		log.Fatal("MTSU_BASE_PATHS is not set")
	}

	basePathsSlice := strings.Split(basePaths, ";;")
	var libraryPaths []Library

	for _, basePath := range basePathsSlice {
		layoutAndPath := strings.SplitN(basePath, ";", 2)
		if len(layoutAndPath) != 2 {
			log.Fatal("MTSU_BASE_PATHS is not set correctly")
		}

		libraryOptionsMatch := libraryOptionsR.MatchString(layoutAndPath[0])
		if !libraryOptionsMatch {
			log.Fatal(layoutAndPath[0], " is not a valid layout in BASE_PATHS. Valid: freeform1, structured2, freeform3 ...")
		}
		if layoutAndPath[1] == "" {
			log.Fatal("Paths in MTSU_BASE_PATHS cannot be empty")
		}
		if _, err := os.Stat(layoutAndPath[1]); os.IsNotExist(err) {
			log.Fatal("Path in MTSU_BASE_PATHS not found: ", layoutAndPath[1])
		}

		libraryOptions := libraryOptionsR.FindStringSubmatch(layoutAndPath[0])
		id, _ := strconv.ParseInt(libraryOptions[2], 10, 32)

		libraryPaths = append(libraryPaths, Library{
			ID:     int32(id),
			Path:   layoutAndPath[1],
			Layout: libraryOptions[1],
		})
	}

	return libraryPaths
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

func GetAddress() string {
	value := os.Getenv("MTSU_HOSTNAME")
	if value == "" {
		return "localhost"
	}
	return value
}

func GetPort() string {
	value := os.Getenv("MTSU_PORT")
	if value == "" {
		return "5050"
	}
	return value
}

func CacheServerDisabled() bool {
	value := os.Getenv("MTSU_DISABLE_CACHE_SERVER")
	if value == "true" {
		return true
	}
	return false
}

func RegistrationsEnabled() bool {
	value := os.Getenv("MTSU_REGISTRATIONS")
	if value == "true" {
		return true
	}
	return false
}

func CurrentVisibility() string {
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

func RestrictedPassphrase() string {
	value := os.Getenv("MTSU_RESTRICTED_PASSPHRASE")
	if value == "" {
		if CurrentVisibility() == Restricted {
			log.Error("MTSU_RESTRICTED_PASSPHRASE is not set. Defaulting to 's3cr3t'.")
		}
		return "s3cr3t"
	}
	return value
}

func JWTSecret() string {
	value := os.Getenv("MTSU_JWT_SECRET")
	if value == "" {
		log.Error("MTSU_JWT_SECRET is not set. An unsecure secret will be used instead. DO NOT USE IN PRODUCTION.")
		return "iugnrg8o9347ghjmloi2jhbaw8723hjdbjnwq"
	}
	return value
}

func RelativePath(basePath string, fullPath string) string {
	return strings.Replace(filepath.ToSlash(fullPath), filepath.ToSlash(basePath)+"/", "", 1)
}

func BuiltPath(base string, pathParts ...string) string {
	if len(pathParts) == 0 {
		return base
	}
	basePath := []string{base}
	pathSlice := append(basePath, pathParts...)
	return filepath.ToSlash(path.Join(pathSlice...))
}

func BuildLibraryPath(libraryPath string, pathParts ...string) string {
	return BuiltPath(libraryPath, pathParts...)
}

func BuildDataPath(pathParts ...string) string {
	return BuiltPath(os.Getenv("MTSU_DATA_PATH"), pathParts...)
}

func BuildCachePath(pathParts ...string) string {
	return BuiltPath(os.Getenv("MTSU_DATA_PATH"), append([]string{"cache"}, pathParts...)...)
}
