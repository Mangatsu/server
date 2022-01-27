package library

import (
	"errors"
	"github.com/Luukuton/Mangatsu/internal/config"
	"github.com/Luukuton/Mangatsu/pkg/db"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func walk(libraryPath string, libraryID int32) fs.WalkDirFunc {
	return func(s string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !ArchiveExtensions.MatchString(d.Name()) {
			return nil
		}

		s = filepath.ToSlash(s)
		relativePath := config.RelativePath(libraryPath, s)

		// Skip if already in database
		if db.ArchivePathFound(relativePath) {
			log.Debug("Skipping: ", d.Name())
			return nil
		}

		n := strings.LastIndex(d.Name(), path.Ext(d.Name()))
		title := d.Name()[:n]

		err = db.NewGallery(relativePath, libraryID, title)
		if err != nil {
			log.Error("Error adding:", err)
			return err
		}

		log.Debug("Added: ", s)

		return err
	}
}

func ScanArchives(full bool) {
	// TODO: full scan
	libraries := db.GetOnlyLibraries()
	for _, library := range libraries {
		err := filepath.WalkDir(library.Path, walk(library.Path, library.ID))
		if err != nil {
			log.Error("Error scanning dir: ", err)
			return
		}
	}
}

func PathExists(pathTo string) bool {
	_, err := os.OpenFile(pathTo, os.O_RDONLY, 0444)
	return !errors.Is(err, os.ErrNotExist)
}
