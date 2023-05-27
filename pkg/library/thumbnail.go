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
	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/mholt/archiver/v4"
	log "github.com/sirupsen/logrus"
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
		log.Error("Could not get libraries for thumbnail generation: ", err)
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
			log.Error("Couldn't create thumbnail dir: ", err)
			return
		}
	}

	fsys, err := archiver.FileSystem(nil, archivePath)
	if err != nil {
		log.Error("Error reading '", archivePath, "' on trying to generate thumbnails")
		return
	}

	if dir, ok := fsys.(fs.ReadDirFile); ok {
		entries, err := dir.ReadDir(0)
		if err != nil {
			log.Error("Error reading dir: ", err)
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
					log.Error("Couldn't create inner thumbnail dir: ", err)
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
			log.Error("Couldn't generate thumbnail for: ", d.Name())
			return err
		}

		if onlyCover {
			webpName := imgName + ".webp"
			log.Info("Cover thumbnail generated: ", webpName)

			if err := db.SetThumbnail(galleryUUID, webpName); err != nil {
				log.Error("Couldn't save cover thumbnail to db: ", err)
				return err
			}
			return errors.New("terminate walk")
		}
		generatedCount++

		return nil
	})
	if err != nil && err.Error() != "terminate walk" {
		log.Debug("Error walking dir: ", err)
	}
	if !onlyCover {
		log.Infof("%d non-cover thumbnails generated for gallery: %s", generatedCount, galleryUUID)
	}
}

// generateThumbnail generates a thumbnail for a given image and saves it to cache as webp.
func generateThumbnail(galleryUUID string, thumbnailPath string, imgBytes []byte, large bool) error {
	srcImage, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		log.Debugf("Could not decode img %s. Message: %s", thumbnailPath, err)
		return err
	}

	width := 250
	if large {
		width = 500
	}

	dstImage := imaging.Resize(srcImage, width, 0, imaging.Lanczos)

	var buf bytes.Buffer
	if err = webp.Encode(&buf, dstImage, &webp.Options{Lossless: false, Quality: 75}); err != nil {
		log.Debugf("Could not encode img %s. Message: %s", thumbnailPath, err)
		return err
	}

	// webp
	err = os.WriteFile(config.BuildCachePath("thumbnails", galleryUUID, thumbnailPath+".webp"), buf.Bytes(), 0666)
	if err != nil {
		log.Debugf("Couldn't write thumbnail %s. Message: %s", thumbnailPath, err)
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
