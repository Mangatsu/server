package metadata

import (
	"bufio"
	"github.com/Mangatsu/server/pkg/types/model"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

// ParseHath parses given text file. Input file is expected to be in the H@H (Hath) format (galleryinfo.txt).
func ParseHath(filePath string) (model.Gallery, []model.Tag, error) {
	file, err := os.Open(filePath)
	gallery := model.Gallery{}
	var tags []model.Tag

	if err != nil {
		return gallery, nil, err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error(err)
		}
	}(file)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "Title:") {
			title := strings.TrimSpace(strings.TrimPrefix(line, "Title:"))
			gallery.TitleNative = &title
		}

		if strings.HasPrefix(line, "Tags:") {
			tagsList := strings.Split(strings.TrimSpace(strings.TrimPrefix(line, "Tags:")), ",")
			for _, tag := range tagsList {
				splitTag := strings.Split(tag, ":")
				if len(splitTag) == 2 {
					tags = append(tags, model.Tag{
						Namespace: strings.TrimSpace(splitTag[0]),
						Name:      strings.TrimSpace(splitTag[1]),
					})
				} else if len(splitTag) == 1 {
					tags = append(tags, model.Tag{
						Namespace: "other",
						Name:      strings.TrimSpace(splitTag[0]),
					})
				}

			}
		}
	}

	return gallery, tags, nil
}
