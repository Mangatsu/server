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
	"github.com/chai2010/webp"
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
				ReadArchiveImages(fullPath, gallery.UUID, onlyCover)
			}()
		}
	}
}

// ReadArchiveImages reads given archive ands generates thumbnails for found images.
// If onlyCover is true, only covers are generated and the name of cover is saved to db, otherwise covers are generated.
func ReadArchiveImages(archivePath string, galleryUUID string, onlyCover bool) {
	galleryThumbnailPath := config.BuildCachePath("thumbnails", galleryUUID)
	if !utils.PathExists(galleryThumbnailPath) {
		err := os.Mkdir(galleryThumbnailPath, os.ModePerm)

		if err != nil && !errors.Is(err, os.ErrExist) {
			log.Z.Error("could not create thumbnail dir",
				zap.String("path", galleryThumbnailPath),
				zap.String("err", err.Error()))

			return
		}
	}

	filesystem, err := archiver.FileSystem(nil, archivePath)
	if err != nil {
		log.Z.Error("failed to read path on trying to generate thumbnails",
			zap.String("path", archivePath),
			zap.String("err", err.Error()))

		return
	}

	if dir, ok := filesystem.(fs.ReadDirFile); ok {
		entries, err := dir.ReadDir(0)
		if err != nil {
			log.Z.Error("failed to read dir", zap.String("err", err.Error()))
			return
		}
		for _, e := range entries {
			fmt.Println(e.Name())
		}
	}

	cover := true
	generatedCount := 0
	err = fs.WalkDir(filesystem, ".", func(s string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if s == "." || s == ".." {
			return nil
		}

		if d.IsDir() {
			cacheInnerDir := config.BuildCachePath("thumbnails", galleryUUID, d.Name())

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

		if !onlyCover && cover {
			cover = false
			return nil
		}

		imgExtension := path.Ext(s)
		n := strings.LastIndex(s, imgExtension)
		imgName := s[:n]

		content, err := ReadAll(filesystem, s)
		if err != nil {
			return err
		}

		err = generateThumbnail(galleryUUID, imgName, content, onlyCover)
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

		if onlyCover {
			webpName := imgName + ".webp"
			log.Z.Info("cover thumbnail generated", zap.String("img", webpName))
			cache.ProcessingStatusCache.AddThumbnailGeneratedCover()

			if err := db.SetThumbnail(galleryUUID, webpName); err != nil {
				log.Z.Error("could not save cover thumbnail to db",
					zap.String("uuid", galleryUUID),
					zap.String("name", webpName),
					zap.String("err", err.Error()))
				return err
			}
			return errors.New("terminate walk")
		}
		cache.ProcessingStatusCache.AddThumbnailGeneratedPage()
		generatedCount++

		return nil
	})
	if err != nil && err.Error() != "terminate walk" {
		log.Z.Debug("failed to walk the dir when generating thumbnails", zap.String("err", err.Error()))
	}
	if !onlyCover {
		err := db.SetPageThumbnails(galleryUUID, int32(generatedCount))
		if err != nil {
			log.Z.Error("could not save page thumbnails status to db",
				zap.String("uuid", galleryUUID),
				zap.String("err", err.Error()))
		}

		log.Z.Info("non-cover thumbnails generated for gallery",
			zap.String("uuid", galleryUUID),
			zap.Int("count", generatedCount))
	}
}

// generateThumbnail generates a thumbnail for a given image and saves it to cache as webp.
func generateThumbnail(galleryUUID string, thumbnailPath string, imgBytes []byte, large bool) error {
	srcImage, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		log.Z.Debug("could not decode img",
			zap.String("path", thumbnailPath),
			zap.String("err", err.Error()))
		return err
	}

	width := 250
	if large {
		width = 500
	}

	dstImage := imaging.Resize(srcImage, width, 0, imaging.Lanczos)

	var buf bytes.Buffer
	if err = webp.Encode(&buf, dstImage, &webp.Options{Lossless: false, Quality: 75}); err != nil {
		log.Z.Debug("could not encode img",
			zap.String("path", thumbnailPath),
			zap.String("err", err.Error()))
		return err
	}

	// webp
	err = os.WriteFile(config.BuildCachePath("thumbnails", galleryUUID, thumbnailPath+".webp"), buf.Bytes(), 0666)
	if err != nil {
		log.Z.Debug("could not write thumbnail",
			zap.String("path", thumbnailPath),
			zap.String("err", err.Error()))
	}

	// TODO: test how long it takes to generate webp thumbnails compared to jpg + size differences
	//newImage, _ := os.Create("../cache/thumbnails/" + galleryUUID + "/" + name)
	//defer func(newImage *os.File) {
	//	err := newImage.Close()
	//	if err != nil {
	//		log.Error("Closing thumbnail: ", err)
	//	}
	//}(newImage)
	//err = png.Encode(newImage, dstImage)
	//if err != nil {
	//	log.Error("Writing thumbnail: ", err)
	//	return
	//}

	return err
}
