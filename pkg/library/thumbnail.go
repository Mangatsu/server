package library

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/log"
	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/mholt/archiver/v4"
	"go.uber.org/zap"
)

// GenerateThumbnails generates thumbnails for covers and pages in parallel.
func GenerateThumbnails(pages bool, force bool) {
	// TODO: force / rewrite existing cache
	go thumbnailWalker(true)
	if pages {
		go thumbnailWalker(false)
	}
}

// thumbnailWalker walks through the database and generates thumbnails for covers or pages depending on onlyCover param.
// If force is set to false, already existing directories will be skipped.
func thumbnailWalker(onlyCover bool) {
	libraries, err := db.GetLibraries()
	if err != nil {
		log.Z.Error("could not get libraries for thumbnail generation", zap.String("err", err.Error()))
		return
	}

	for _, library := range libraries {
		for _, gallery := range library.Galleries {
			fullPath := config.BuildLibraryPath(library.Path, gallery.ArchivePath)
			go ReadArchiveImages(fullPath, gallery.UUID, onlyCover)
		}
	}
}

// ReadArchiveImages reads given archive ands generates thumbnails for found images.
// If onlyCover is true, only covers are generated and the name of cover is saved to db, otherwise covers are generated.
func ReadArchiveImages(archivePath string, galleryUUID string, onlyCover bool) {
	galleryThumbnailPath := config.BuildCachePath("thumbnails", galleryUUID)
	if !PathExists(galleryThumbnailPath) {
		err := os.Mkdir(galleryThumbnailPath, os.ModePerm)
		if err != nil {
			log.Z.Error("could not create thumbnail dir",
				zap.String("path", galleryThumbnailPath),
				zap.String("err", err.Error()))
			return
		}
	}

	fsys, err := archiver.FileSystem(nil, archivePath)
	if err != nil {
		log.Z.Error("failed to read path on trying to generate thumbnails",
			zap.String("path", archivePath),
			zap.String("err", err.Error()))
		return
	}

	if dir, ok := fsys.(fs.ReadDirFile); ok {
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
	err = fs.WalkDir(fsys, ".", func(s string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if s == "." || s == ".." {
			return nil
		}

		if d.IsDir() {
			cacheInnerDir := config.BuildCachePath("thumbnails", galleryUUID, d.Name())
			if !PathExists(cacheInnerDir) {
				if err = os.Mkdir(cacheInnerDir, os.ModePerm); err != nil {
					log.Z.Error("could not create inner thumbnail dir",
						zap.String("path", cacheInnerDir),
						zap.String("err", err.Error()))
					return err
				}
			}
			return nil
		}

		if !ImageExtensions.MatchString(d.Name()) {
			return nil
		}
		if !onlyCover && cover {
			cover = false
			return nil
		}

		imgExtension := path.Ext(s)
		n := strings.LastIndex(s, imgExtension)
		imgName := s[:n]

		content, err := ReadAll(fsys, s)
		if err != nil {
			return err
		}

		err = generateThumbnail(galleryUUID, imgName, content, onlyCover)
		if err != nil {
			log.Z.Error("could not create thumbnail",
				zap.String("uuid", galleryUUID),
				zap.String("name", d.Name()),
				zap.String("err", err.Error()))
			return err
		}

		if onlyCover {
			webpName := imgName + ".webp"
			log.Z.Info("cover thumbnail generated", zap.String("img", webpName))

			if err := db.SetThumbnail(galleryUUID, webpName); err != nil {
				log.Z.Error("could not save cover thumbnail to db",
					zap.String("uuid", galleryUUID),
					zap.String("name", webpName),
					zap.String("err", err.Error()))
				return err
			}
			return errors.New("terminate walk")
		}
		generatedCount++

		return nil
	})
	if err != nil && err.Error() != "terminate walk" {
		log.Z.Debug("failed to walk the dir when generating thumbnails", zap.String("err", err.Error()))
	}
	if !onlyCover {
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
