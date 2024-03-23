package utils

import (
	"errors"
	"github.com/Mangatsu/server/pkg/constants"
	"github.com/Mangatsu/server/pkg/log"
	"github.com/mholt/archiver/v4"
	"go.uber.org/zap"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// UniversalExtract extracts media files from zip, cbz, rar, cbr, tar (all its variants) archives.
// Plain directories without compression are also supported. For PDF files, use ExtractPDF.
func UniversalExtract(dst string, archivePath string) ([]string, int) {
	fsys, err := archiver.FileSystem(nil, archivePath)
	if err != nil {
		log.Z.Error("failed to open an archive",
			zap.String("path", archivePath),
			zap.String("err", err.Error()))
		return nil, 0
	}

	if err = os.Mkdir(dst, os.ModePerm); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			log.Z.Error("failed to create a dir for gallery",
				zap.String("path", dst),
				zap.String("err", err.Error()))
			return nil, 0

		}
	}

	var files []string
	count := 0

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
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}

		if !constants.ImageExtensions.MatchString(d.Name()) {
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
			log.Z.Error("failed to copy file",
				zap.String("dstFile", dstFile.Name()),
				zap.String("err", err.Error()))
		}

		if err = dstFile.Close(); err != nil {
			return err
		}
		if err = fileInArchive.Close(); err != nil {
			return err
		}

		files = append(files, s)
		count += 1
		return nil
	})
	if err != nil {
		log.Z.Debug("failed to walk dir when copying archive", zap.String("err", err.Error()))
		return nil, 0
	}

	return files, count
}

func ExtractPDF() {
	// TODO: Add support for PDF files, Probably with https://github.com/gen2brain/go-fitz
}
