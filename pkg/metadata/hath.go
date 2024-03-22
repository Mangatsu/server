package metadata

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/Mangatsu/server/pkg/types/sqlite/model"
)

// ParseHath parses given text file. Input file is expected to be in the H@H (Hath) format (galleryinfo.txt).
func ParseHath(metaPath string, metaData []byte, internal bool) (model.Gallery, []model.Tag, model.Reference, error) {
	gallery := model.Gallery{}
	reference := model.Reference{
		MetaPath:     &metaPath,
		MetaInternal: internal,
		Urls:         nil,
	}
	var tags []model.Tag

	buffer := bytes.NewBuffer(metaData)
	scanner := bufio.NewScanner(buffer)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "Title:") {
			title := strings.TrimSpace(strings.TrimPrefix(line, "Title:"))
			gallery.TitleNative = &title
			continue
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

	return gallery, tags, reference, nil
}
