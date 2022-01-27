package library

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Luukuton/Mangatsu/internal/config"
	"github.com/Luukuton/Mangatsu/pkg/db"
	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/mholt/archiver/v4"
	log "github.com/sirupsen/logrus"
	"image"
	"io/fs"
	"os"
	"path"
	"strings"
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
	libraries := db.GetLibraries()
	for _, library := range libraries {
		for _, gallery := range library.Galleries {
			galleryThumbnailPath := config.BuildCachePath("thumbnails", gallery.UUID)
			if !PathExists(galleryThumbnailPath) {
				err := os.Mkdir(galleryThumbnailPath, os.ModePerm)
				if err != nil {
					log.Error(err)
					continue
				}
			}

			log.Info("Generating thumbnails for gallery: " + gallery.UUID)

			fullPath := config.BuildLibraryPath(library.Path, gallery.ArchivePath)

			readArchiveImages(fullPath, gallery.UUID, onlyCover)
		}
	}
}

// readArchiveImages reads given archive ands generates thumbnails for found images.
// If onlyCover is true, only covers are generated and the name of cover is saved to db, otherwise covers are generated.
func readArchiveImages(archivePath string, galleryUUID string, onlyCover bool) {
	fsys, err := archiver.FileSystem(archivePath)
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
				log.Info("Creating thumbnail dir: " + cacheInnerDir)
				err = os.Mkdir(cacheInnerDir, os.ModePerm)
				if err != nil {
					log.Error("Error creating inner dir for cache", err)
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

		content, err := readAll(fsys, s)
		if err != nil {
			return err
		}

		err = generateThumbnail(galleryUUID, imgName, content, onlyCover)
		if err != nil {
			log.Error("Error generating thumbnail for", d.Name())
			return err
		}

		if onlyCover {
			if err == nil {
				log.Info("Cover thumbnail generated: ", d.Name())
				db.SetThumbnail(galleryUUID, imgName+".webp")
				if err != nil {
					log.Error("Error saving thumbnail to db")
				}
			}
			return errors.New("terminate walk")
		}

		return nil
	})
	if err != nil && err.Error() != "terminate walk" {
		log.Error("Error walking dir: ", err)
	}
}

// generateThumbnail generates a thumbnail for a given image and saves it to cache as webp.
func generateThumbnail(galleryUUID string, thumbnailPath string, imgBytes []byte, large bool) error {
	srcImage, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		log.Error("Decode: ", err)
		return err
	}

	width := 250
	if large {
		width = 500
	}

	dstImage := imaging.Resize(srcImage, width, 0, imaging.Lanczos)

	var buf bytes.Buffer
	if err = webp.Encode(&buf, dstImage, &webp.Options{Lossless: false, Quality: 75}); err != nil {
		log.Error("Encode: ", err)
		return err
	}

	// webp
	err = os.WriteFile(config.BuildCachePath("thumbnails", galleryUUID, thumbnailPath+".webp"), buf.Bytes(), 0666)
	if err != nil {
		log.Error("Writing thumbnail: ", err)
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
