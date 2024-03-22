package metadata

import (
	"bufio"
	"bytes"
	"regexp"
	"strconv"
	"strings"

	"github.com/Mangatsu/server/pkg/log"
	"github.com/Mangatsu/server/pkg/types/sqlite/model"
)

var exhURLRegex = regexp.MustCompile(`https://\w+\.\w+/g/(\d+)/[a-z0-9]+`)
var lengthRegex = regexp.MustCompile(`Length:\s*(\d+)`)
var sizeRegex = regexp.MustCompile(`File Size:\s*(\d+(?:\.\d+)?)`)

// ParseEHDL parses given text file. Input file is expected to be in the H@H (Hath) format (galleryinfo.txt).
// Input file is expected to be in the E-Hentai-Downloader format (info.txt).
func ParseEHDL(metaPath string, metaData []byte, internal bool) (model.Gallery, []model.Tag, model.Reference, error) {
	gallery := model.Gallery{}
	reference := model.Reference{
		MetaPath:     &metaPath,
		MetaInternal: internal,
		Urls:         nil,
	}
	var tags []model.Tag

	buffer := bytes.NewBuffer(metaData)
	scanner := bufio.NewScanner(buffer)
	lineNumber := -1

	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++
		if len(line) == 0 {
			continue
		}

		switch lineNumber {
		case 0:
			gallery.Title = strings.TrimSpace(line)
			continue
		case 1:
			titleNative := strings.TrimSpace(line)
			gallery.TitleNative = &titleNative
			continue
		case 2:
			// https://example.org/g/999999/f2f2f2f2f2/
			capture := exhURLRegex.FindStringSubmatch(line)
			if len(capture) < 3 {
				continue
			}
			exhGid, err := strconv.ParseInt(capture[2], 10, 32)
			if err != nil {
				log.S.Debug("failed to parse exhGid", "err", err.Error(), "exhGid", exhGid)
				continue
			}
			exhGidInt32 := int32(exhGid)

			reference.ExhGid = &exhGidInt32
			reference.ExhToken = &capture[1]
			continue
		}

		if strings.HasPrefix(line, "Category:") {
			category := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "Category:")))
			gallery.Category = &category
			continue
		}

		if strings.HasPrefix(line, "Language:") {
			language := strings.TrimSpace(strings.TrimPrefix(line, "Language:"))
			gallery.Language = &language
			continue
		}

		if strings.HasPrefix(line, "Length:") {
			capture := lengthRegex.FindStringSubmatch(line)
			if len(capture) < 2 {
				continue
			}
			length, err := strconv.ParseInt(capture[1], 10, 32)
			if err != nil {
				log.S.Debug("failed to parse length", "err", err.Error(), "length", length)
				continue
			}
			lengthInt32 := int32(length)
			gallery.ImageCount = &lengthInt32
			continue
		}

		if strings.HasPrefix(line, "File Size:") {
			capture := sizeRegex.FindStringSubmatch(line)
			if len(capture) < 2 {
				continue
			}
			size, err := strconv.ParseFloat(capture[1], 32)
			if err != nil {
				log.S.Debug("failed to parse size", "err", err.Error(), "size", size)
				continue
			}
			sizeInt32 := int32(size * 1000 * 1000)
			gallery.ArchiveSize = &sizeInt32
			continue
		}

		if strings.HasPrefix(line, "> ") {
			tagsList := strings.Split(strings.TrimSpace(strings.TrimPrefix(line, "> ")), ":")
			if len(tagsList) != 2 {
				names := strings.Split(tagsList[1], ",")
				for _, name := range names {
					tags = append(tags, model.Tag{
						Namespace: tagsList[0],
						Name:      strings.TrimSpace(name),
					})
				}
			}
		}
	}

	return gallery, tags, reference, nil
}
