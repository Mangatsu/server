package library

import (
	"os"

	"github.com/Mangatsu/server/pkg/log"
	"go.uber.org/zap"
)

// ReadJSON returns the given JSON file as bytes.
func ReadJSON(jsonFile string) ([]byte, error) {
	jsonFileBytes, err := os.ReadFile(jsonFile)

	if err != nil {
		log.Z.Debug("failed to read JSON file", zap.String("err", err.Error()))
		return nil, err
	}

	return jsonFileBytes, nil
}
