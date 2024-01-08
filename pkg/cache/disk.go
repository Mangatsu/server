package cache

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/log"
	"github.com/Mangatsu/server/pkg/utils"
	"github.com/facette/natsort"
	"go.uber.org/zap"
)

// InitPhysicalCache initializes the physical cache directories.
func InitPhysicalCache() {
	cachePath := config.BuildCachePath()
	if !utils.PathExists(cachePath) {
		err := os.Mkdir(cachePath, os.ModePerm)
		if err != nil {
			log.Z.Error("failed to create cache dir",
				zap.String("path", cachePath),
				zap.String("err", err.Error()))
		}
	}

	thumbnailsPath := config.BuildCachePath("thumbnails")
	if !utils.PathExists(thumbnailsPath) {
		err := os.Mkdir(thumbnailsPath, os.ModePerm)
		if err != nil {
			log.Z.Error("failed to create thumbnails cache dir",
				zap.String("path", thumbnailsPath),
				zap.String("err", err.Error()))
		}
	}
}

// readPhysicalCache reads the physical cache from the disk and returns the list of files and the number of files.
func readPhysicalCache(dst string, uuid string) ([]string, int) {
	var files []string
	count := 0

	cacheWalk := func(s string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Z.Error("failed to walk cache dir",
				zap.String("name", d.Name()),
				zap.String("err", err.Error()))
			return err
		}
		if d.IsDir() {
			return nil
		}

		// ReplaceAll ensures that the path is correct: cache/uuid/<arbitrary/path/image.png>
		files = append(files, strings.ReplaceAll(filepath.ToSlash(s), config.BuildCachePath(uuid)+"/", ""))
		count += 1
		return nil
	}

	err := filepath.WalkDir(dst, cacheWalk)
	if err != nil {
		log.Z.Error("failed to walk cache dir",
			zap.String("dst", dst),
			zap.String("err", err.Error()))
		return nil, 0
	}

	return files, count
}

// extractGallery extracts the gallery from the archive and returns the list of files and the number of files.
func extractGallery(archivePath string, uuid string) ([]string, int) {
	dst := config.BuildCachePath(uuid)
	if _, err := os.Stat(dst); errors.Is(err, fs.ErrNotExist) {
		return utils.UniversalExtract(dst, archivePath)
	}

	files, count := readPhysicalCache(dst, uuid)
	if count == 0 {
		err := os.Remove(dst)
		if err != nil {
			log.Z.Debug("removing empty cache dir failed",
				zap.String("dst", dst),
				zap.String("err", err.Error()))
			return nil, 0
		}

		return utils.UniversalExtract(dst, archivePath)
	}
	natsort.Sort(files)

	return files, count
}
