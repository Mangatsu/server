package library

import (
	"errors"
	"github.com/Mangatsu/server/internal/config"
	"github.com/mholt/archiver/v4"
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

func readAll(fsys fs.FS, filename string) ([]byte, error) {
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

// ReadArchiveInternalMeta reads the internal metadata (info.json or info.txt) from an archive
func ReadArchiveInternalMeta(archivePath string) ([]byte, string) {
	fsys, err := archiver.FileSystem(archivePath)
	if err != nil {
		log.Error("Error opening archive: ", err)
		return nil, ""
	}

	var content []byte
	var filename string

	err = fs.WalkDir(fsys, ".", func(s string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// TODO: Change check to this when txt metafiles are supported: MetaExtensions.MatchString(d.Name())
		if d.Name() == "info.json" {
			content, err = readAll(fsys, s)
			if err != nil {
				return err
			}

			filename = s
			return errors.New("terminate walk")
		}
		return nil
	})

	return content, filename
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

	return files, count
}
