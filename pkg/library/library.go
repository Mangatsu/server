package library

import (
	"github.com/Mangatsu/server/pkg/log"
	"go.uber.org/zap"
	"io"
	"io/fs"
)

func closeFile(f interface{ Close() error }) {
	err := f.Close()
	if err != nil {
		log.Z.Debug("failed to close file", zap.String("err", err.Error()))
		return
	}
}

func ReadAll(filesystem fs.FS, filename string) ([]byte, error) {
	archive, err := filesystem.Open(filename)
	if err != nil {
		return nil, err
	}

	defer closeFile(archive)

	return io.ReadAll(archive)
}
