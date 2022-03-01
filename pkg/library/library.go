package library

import (
	"github.com/Mangatsu/server/internal/config"
	"github.com/facette/natsort"
	log "github.com/sirupsen/logrus"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func closeFile(f interface{ Close() error }) {
	err := f.Close()
	if err != nil {
		log.Error(err)
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
			log.Error("Error in walk: ", err)
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
		log.Error("Error walking dir: ", err)
		return nil, 0
	}

	return files, count
}

func ReadGallery(archivePath string, uuid string) ([]string, int) {
	dst := config.BuildCachePath(uuid)
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		return UniversalExtract(dst, archivePath)
	}

	files, count := readCache(dst, uuid)
	if count == 0 {
		err := os.Remove(dst)
		if err != nil {
			return nil, 0
		}

		return UniversalExtract(dst, archivePath)
	}
	natsort.Sort(files)

	return files, count
}
