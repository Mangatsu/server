package library

import (
	"github.com/Mangatsu/server/internal/config"
	"github.com/mholt/archiver/v4"
	log "github.com/sirupsen/logrus"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func InitCache() {
	cachePath := config.BuildCachePath()
	if !PathExists(cachePath) {
		err := os.Mkdir(cachePath, os.ModePerm)
		if err != nil {
			log.Error(err)
		}
	}

	thumbnailsPath := config.BuildCachePath("thumbnails")
	if !PathExists(thumbnailsPath) {
		err := os.Mkdir(thumbnailsPath, os.ModePerm)
		if err != nil {
			log.Error(err)
		}
	}
}

func ExtractPDF() {
	// TODO: Add support for PDF files, Probably with https://github.com/gen2brain/go-fitz
}

func Extract7z() {
	// TODO: add support for 7z compression. Probably with https://github.com/bodgit/sevenzip
}

// UniversalExtract extracts media files from zip, cbz, rar, cbr, tar (all its variants) archives.
// Plain directories without compression are also supported. For 7zip and PDF see Extract7z and ExtractPDF respectively.
func UniversalExtract(dst string, archivePath string) ([]string, int32) {
	fsys, err := archiver.FileSystem(archivePath)
	if err != nil {
		log.Error("Error opening archive: ", err)
		return nil, 0
	}

	err = os.Mkdir(dst, os.ModePerm)
	if err != nil {
		log.Error(err)
		return nil, 0
	}

	var files []string
	count := int32(0)

	err = fs.WalkDir(fsys, ".", func(s string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, s)

		if s == "." || s == ".." {
			return nil
		}

		if d.IsDir() {
			err = os.Mkdir(dstPath, os.ModePerm)
			return err
		}

		if !ImageExtensions.MatchString(d.Name()) {
			return nil
		}

		//if !strings.HasPrefix(dstPath, filepath.Clean(dst)+string(os.PathSeparator)) {
		//	log.Error("Invalid file path: ", dstPath)
		//	return nil
		//}

		fileInArchive, err := fsys.Open(s)
		if err != nil {
			return err
		}

		dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			log.Error(err)
		}
		err = dstFile.Close()
		err = fileInArchive.Close()
		if err != nil {
			return err
		}

		files = append(files, s)
		count += 1
		return nil
	})
	if err != nil {
		log.Error("Error walking dir: ", err)
		return nil, 0
	}

	return files, count
}
