package library

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/log"
	"github.com/facette/natsort"
	"go.uber.org/zap"
)

func closeFile(f interface{ Close() error }) {
	err := f.Close()
	if err != nil {
		log.Z.Debug("failed to close file", zap.String("err", err.Error()))
		return
	}
}

func ReadAll(fsys fs.FS, filename string) ([]byte, error) {
	archive, err := fsys.Open(filename)
	if err != nil {
		return nil, err
	}

	defer closeFile(archive)

	return io.ReadAll(archive)
}

func readCache(dst string, uuid string) ([]string, int) {
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

func ReadGallery(archivePath string, uuid string) ([]string, int) {
	dst := config.BuildCachePath(uuid)
	if _, err := os.Stat(dst); errors.Is(err, fs.ErrNotExist) {
		return UniversalExtract(dst, archivePath)
	}

	files, count := readCache(dst, uuid)
	if count == 0 {
		err := os.Remove(dst)
		if err != nil {
			log.Z.Debug("removing empty cache dir failed",
				zap.String("dst", dst),
				zap.String("err", err.Error()))
			return nil, 0
		}

		return UniversalExtract(dst, archivePath)
	}
	natsort.Sort(files)

	return files, count
}
