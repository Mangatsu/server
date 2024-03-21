package library

import (
	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/cache"
	"github.com/Mangatsu/server/pkg/constants"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/log"
	"github.com/Mangatsu/server/pkg/utils"
	"github.com/mholt/archiver/v4"
	"go.uber.org/zap"
	"io/fs"
	"path"
	"path/filepath"
	"strings"
)

func countImages(archivePath string) (uint64, error) {
	filesystem, err := archiver.FileSystem(nil, archivePath)
	var fileCount uint64

	err = fs.WalkDir(filesystem, ".", func(s string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && constants.ImageExtensions.MatchString(d.Name()) {
			fileCount++
		}

		return nil
	})

	if err != nil {
		log.Z.Error("could not count files in archive", zap.String("path", archivePath), zap.String("err", err.Error()))
		return 0, err
	}

	return fileCount, nil
}

func walk(libraryPath string, libraryID int32, libraryLayout config.Layout) fs.WalkDirFunc {
	return func(s string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		isImage := constants.ImageExtensions.MatchString(d.Name())
		isArchive := constants.ArchiveExtensions.MatchString(d.Name())
		if !isArchive && !isImage {
			return nil
		}

		s = filepath.ToSlash(s)
		relativePath := config.RelativePath(libraryPath, s)
		fullPath := config.BuildLibraryPath(libraryPath, relativePath)

		// If an image is found, the parent dir will be considered as a gallery.
		if isImage {
			relativePath = path.Dir(relativePath)
		}

		// Skip if already in database
		foundGallery := db.ArchivePathFound(relativePath)
		if foundGallery != nil {
			log.Z.Debug("skipping archive already in db", zap.String("name", d.Name()))

			cache.ProcessingStatusCache.AddScanSkippedGallery(foundGallery[0].UUID)
			return nil
		}

		// Series name from the dir name if Structured layout
		series := ""
		if libraryLayout == config.Structured {
			dirs := strings.SplitN(relativePath, "/", 2)
			series = dirs[0]
		}

		var title string
		var size int64

		if isImage {
			title = path.Base(relativePath)

			if size, err = utils.DirSize(fullPath); err != nil {
				log.Z.Error("failed to get dir size", zap.String("path", fullPath), zap.String("err", err.Error()))
			}
		} else {
			n := strings.LastIndex(d.Name(), path.Ext(d.Name()))
			title = d.Name()[:n]

			if size, err = utils.FileSize(fullPath); err != nil {
				log.Z.Error("failed to get file size", zap.String("path", fullPath), zap.String("err", err.Error()))
			}
		}

		imageCount, err := countImages(fullPath)
		if err != nil {
			log.Z.Error("failed to count images", zap.String("path", fullPath), zap.String("err", err.Error()))
		}

		uuid, err := db.NewGallery(relativePath, libraryID, title, series, size, imageCount)

		if err != nil {
			log.Z.Error("failed to add gallery to db",
				zap.String("path", relativePath),
				zap.String("err", err.Error()))

			cache.ProcessingStatusCache.AddScanError(relativePath, err.Error(), map[string]string{
				"libraryID": string(libraryID),
				"title":     title,
			})

		} else {
			// Generates cover thumbnail
			go ReadArchiveImages(fullPath, uuid, true)

			log.Z.Debug("added gallery", zap.String("path", relativePath))

			cache.ProcessingStatusCache.AddScanFoundGallery(uuid)
		}

		if isImage {
			return fs.SkipDir
		}

		return err
	}
}

func ScanArchives() {
	// TODO: Quick scan by only checking directories that have been modified.
	// Not too important as current implementation is pretty fast already.
	libraries, err := db.GetOnlyLibraries()
	if err != nil {
		log.Z.Error("failed to find libraries to scan", zap.String("err", err.Error()))

		cache.ProcessingStatusCache.AddScanError("library scan fail", err.Error(), nil)
		return
	}

	cache.ProcessingStatusCache.SetScanRunning(true)
	defer cache.ProcessingStatusCache.SetScanRunning(false)

	for _, library := range libraries {
		err := filepath.WalkDir(library.Path, walk(library.Path, library.ID, config.Layout(library.Layout)))
		if err != nil {
			log.Z.Error("skipping library as an error occurred during scanning",
				zap.String("path", library.Path),
				zap.String("err", err.Error()))

			cache.ProcessingStatusCache.AddScanError(library.Path, err.Error(), nil)
			continue
		}
	}
}
