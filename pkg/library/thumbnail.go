package library

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/cache"
	"github.com/Mangatsu/server/pkg/constants"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/log"
	"github.com/Mangatsu/server/pkg/utils"
	"github.com/disintegration/imaging"
	"github.com/mholt/archiver/v4"
	"go.uber.org/zap"
	"image"
	"io/fs"
	"os"
	"path"
	"strings"
	"sync"
)

// GenerateThumbnails generates thumbnails for covers and pages in parallel. // TODO: ignore generated files or rewrite existing cache option
func GenerateThumbnails(pages bool, force bool) {
	var wg sync.WaitGroup

	cache.ProcessingStatusCache.SetThumbnailsRunning(true)
	coverCount, pageCount, err := db.CountAllImages(false)
	if err != nil {
		log.Z.Error("could not get image count for thumbnail generation", zap.String("err", err.Error()))
	} else {
		log.Z.Info("thumbnail generation started", zap.Int("covers", coverCount), zap.Int("pages", pageCount))
		cache.ProcessingStatusCache.SetTotalCoversAndPages(coverCount, pageCount)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		thumbnailWalker(&wg, true)
	}()

	if pages {
		wg.Add(1)
		go func() {
			defer wg.Done()
			thumbnailWalker(&wg, false)
		}()
	}

	wg.Wait()
	cache.ProcessingStatusCache.SetThumbnailsRunning(false)
}

// thumbnailWalker walks through the database and generates thumbnails for covers or pages depending on onlyCover param.
// If force is set to false, already existing directories will be skipped.
func thumbnailWalker(wg *sync.WaitGroup, onlyCover bool) {
	libraries, err := db.GetLibraries()
	if err != nil {
		log.Z.Error("could not get libraries for thumbnail generation", zap.String("err", err.Error()))
		return
	}

	for _, library := range libraries {
		for _, gallery := range library.Galleries {
			wg.Add(1)

			fullPath := config.BuildLibraryPath(library.Path, gallery.ArchivePath)
			gallery := gallery // Loop variables captured by 'func' literals in 'go' statements might have unexpected values

			go func() {
				defer wg.Done()
				if onlyCover {
					GenerateCoverThumbnail(fullPath, gallery.UUID)
				} else {
					GeneratePageThumbnails(fullPath, gallery.UUID)
				}
			}()
		}
	}
}

// readArchiveImages reads given archive ands returns the filesystem of the archive.
func readArchiveImages(archivePath string, galleryUUID string) *fs.FS {
	galleryThumbnailPath := config.BuildCachePath("thumbnails", galleryUUID)
	if !utils.PathExists(galleryThumbnailPath) {
		err := os.Mkdir(galleryThumbnailPath, os.ModePerm)

		if err != nil && !errors.Is(err, os.ErrExist) {
			log.Z.Error("could not create thumbnail dir",
				zap.String("path", galleryThumbnailPath),
				zap.String("err", err.Error()))

			return nil
		}
	}

	filesystem, err := archiver.FileSystem(nil, archivePath)
	if err != nil {
		log.Z.Error("failed to read path on trying to generate thumbnails",
			zap.String("path", archivePath),
			zap.String("err", err.Error()))

		return nil
	}

	if dir, ok := filesystem.(fs.ReadDirFile); ok {
		entries, err := dir.ReadDir(0)
		if err != nil {
			log.Z.Error("failed to read dir", zap.String("err", err.Error()))
			return nil
		}
		for _, e := range entries {
			fmt.Println(e.Name())
		}
	}

	return &filesystem
}

// GeneratePageThumbnails generates page thumbnails.
func GeneratePageThumbnails(archivePath string, galleryUUID string) {
	filesystem := readArchiveImages(archivePath, galleryUUID)
	if filesystem == nil {
		return
	}

	generatedCount := 0
	err := fs.WalkDir(*filesystem, ".", func(s string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if s == ".." {
			return nil
		}

		if d.IsDir() {
			cacheInnerDir := config.BuildCachePath("thumbnails", "p", galleryUUID, d.Name())

			if !utils.PathExists(cacheInnerDir) {
				if err = os.Mkdir(cacheInnerDir, os.ModePerm); err != nil {
					log.Z.Error("could not create inner thumbnail dir",
						zap.String("path", cacheInnerDir),
						zap.String("err", err.Error()))

					return err
				}
			}
			return nil
		}

		if !constants.ImageExtensions.MatchString(d.Name()) {
			return nil
		}

		content, err := ReadAll(*filesystem, s)
		if err != nil {
			return err
		}

		// Change the extension to the thumbnail format
		imgExtension := path.Ext(s)
		n := strings.LastIndex(s, imgExtension)
		imgName := s[:n] + "." + string(config.Options.GalleryOptions.ThumbnailFormat)

		err = generateThumbnail(galleryUUID, imgName, content, false)
		if err != nil {
			log.Z.Error("could not create thumbnail",
				zap.String("uuid", galleryUUID),
				zap.String("name", d.Name()),
				zap.String("err", err.Error()))

			cache.ProcessingStatusCache.AddThumbnailError(galleryUUID, err.Error(), map[string]string{
				"name": d.Name(),
			})
			return err
		}

		cache.ProcessingStatusCache.AddThumbnailGeneratedPage()
		generatedCount++

		return nil
	})
	if err != nil && err.Error() != "terminate walk" {
		log.Z.Debug("failed to walk the dir when generating thumbnails", zap.String("err", err.Error()))
	}

	if err = db.SetPageThumbnails(galleryUUID, int32(generatedCount)); err != nil {
		log.Z.Error("could not save page thumbnails status to db",
			zap.String("uuid", galleryUUID),
			zap.String("err", err.Error()))
	}

	log.Z.Info("non-cover thumbnails generated for gallery",
		zap.String("uuid", galleryUUID),
		zap.Int("count", generatedCount))
}

// GenerateCoverThumbnail generates a cover thumbnail.
func GenerateCoverThumbnail(archivePath string, galleryUUID string) {
	filesystem := readArchiveImages(archivePath, galleryUUID)
	if filesystem == nil {
		return
	}

	err := fs.WalkDir(*filesystem, ".", func(s string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if s == ".." || d.IsDir() || !constants.ImageExtensions.MatchString(d.Name()) {
			return nil
		}

		content, err := ReadAll(*filesystem, s)
		if err != nil {
			return err
		}

		imgName := "cover." + string(config.Options.GalleryOptions.ThumbnailFormat)
		err = generateThumbnail(galleryUUID, imgName, content, true)
		if err != nil {
			log.Z.Error("could not create cover thumbnail",
				zap.String("uuid", galleryUUID),
				zap.String("name", d.Name()),
				zap.String("err", err.Error()))

			cache.ProcessingStatusCache.AddThumbnailError(galleryUUID, err.Error(), map[string]string{
				"name": d.Name(),
			})
			return err
		}

		log.Z.Info("cover thumbnail generated", zap.String("img", imgName), zap.String("uuid", galleryUUID))
		cache.ProcessingStatusCache.AddThumbnailGeneratedCover()

		if err := db.SetThumbnail(galleryUUID, imgName); err != nil {
			log.Z.Error("could not save cover thumbnail to db",
				zap.String("uuid", galleryUUID),
				zap.String("name", imgName),
				zap.String("err", err.Error()))
			return err
		}
		return errors.New("terminate walk")
	})

	if err != nil && err.Error() != "terminate walk" {
		log.Z.Debug("failed to walk the dir when generating cover thumbnail", zap.String("err", err.Error()))
	}
}

// generateThumbnail generates a thumbnail for a given image and saves it to cache.
func generateThumbnail(galleryUUID string, imageName string, imgBytes []byte, cover bool) error {
	srcImage, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		log.Z.Debug("could not decode img",
			zap.String("imageName", imageName),
			zap.String("err", err.Error()))
		return err
	}

	var thumbnailPath string
	var width int
	if cover {
		width = 512
		thumbnailPath = config.BuildCachePath("thumbnails", galleryUUID, imageName)
	} else {
		width = 256
		thumbnailPath = config.BuildCachePath("thumbnails", "p", galleryUUID, imageName) // p for pages
	}

	dstImage := imaging.Resize(srcImage, width, 0, imaging.Lanczos)
	buf, err := utils.EncodeImage(dstImage)
	if err != nil {
		log.Z.Debug("could not encode image",
			zap.String("err", err.Error()),
			zap.String("imageName", imageName),
			zap.String("uuid", galleryUUID),
		)
		return err
	}

	err = os.WriteFile(
		thumbnailPath,
		buf.Bytes(),
		0666,
	)
	if err != nil {
		log.Z.Debug("could not write image file",
			zap.String("err", err.Error()),
			zap.String("thumbnailPath", thumbnailPath),
			zap.String("uuid", galleryUUID),
		)
		return err
	}

	return err
}
