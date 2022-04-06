package config

import (
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

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

func RelativePath(basePath string, fullPath string) string {
	return strings.Replace(filepath.ToSlash(fullPath), filepath.ToSlash(basePath)+"/", "", 1)
}

func BuildPath(base string, pathParts ...string) string {
	if len(pathParts) == 0 {
		return base
	}
	basePath := []string{base}
	pathSlice := append(basePath, pathParts...)
	return filepath.ToSlash(path.Join(pathSlice...))
}

func BuildLibraryPath(libraryPath string, pathParts ...string) string {
	return BuildPath(libraryPath, pathParts...)
}

func BuildDataPath(pathParts ...string) string {
	return BuildPath(os.Getenv("MTSU_DATA_PATH"), pathParts...)
}

func BuildCachePath(pathParts ...string) string {
	return BuildPath(os.Getenv("MTSU_DATA_PATH"), append([]string{"cache"}, pathParts...)...)
}
