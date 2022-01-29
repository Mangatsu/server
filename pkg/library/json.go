package library

import (
	log "github.com/sirupsen/logrus"
	"os"
)

// ReadJSON returns the given JSON file as bytes.
func ReadJSON(jsonFile string) ([]byte, error) {
	jsonFileBytes, err := os.ReadFile(jsonFile)

	if err != nil {
		log.Debug("Error in reading json file: ", err)
		return nil, err
	}

	return jsonFileBytes, nil
}
