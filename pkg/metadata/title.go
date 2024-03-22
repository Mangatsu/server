package metadata

import (
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/log"
	"github.com/Mangatsu/server/pkg/types/sqlite/model"
	"go.uber.org/zap"
)

var nameRegex = regexp.MustCompile(
	"(?i)(?:\\(([^([]+)\\))?\\s*(?:\\[([^()[\\]]+)(?:\\(([^()[\\]]+)\\))?])?\\s*([^([]+)\\s*(?:\\(([^([)]+)\\))?\\s*(?:\\[(?:DL版?|digital)])?\\s*(?:\\[([^\\]]+)])?",
)

type TitleMeta struct {
	Released string
	Circle   string
	Artists  []string
	Title    string
	Series   string
	Language string
}

// ParseTitles parses all filenames and titles of the saved galleries in db.
// tryNative tries to preserve the native language (usually Japanese) text.
// overwrite writes over the previous values.
func ParseTitles(tryNative bool, overwrite bool) {
	libraries, err := db.GetLibraries()
	if err != nil {
		log.Z.Error("libraries could not be retrieved to parse titles", zap.String("err", err.Error()))
		return
	}

	for _, library := range libraries {
		for _, gallery := range library.Galleries {
			_, currentTags, err := db.GetTags(gallery.UUID, false)
			currentReference, err := db.GetReference(gallery.UUID)
			if err != nil {
				log.Z.Error("tags could not be retrieved when parsing titles", zap.String("err", err.Error()))
				continue
			}

			hasTitleTranslated := gallery.TitleTranslated != nil
			hasRelease := gallery.Released != nil
			hasSeries := gallery.Series != nil
			hasLanguage := gallery.Language != nil
			hasCircle := containsTag(currentTags, "circle", nil)
			hasArtists := containsTag(currentTags, "artist", nil)

			if !overwrite && hasRelease && hasSeries && hasLanguage && hasCircle && hasArtists {
				continue
			}

			title := gallery.Title
			titleNative := gallery.TitleNative

			filename := filepath.Base(gallery.ArchivePath)
			n := strings.LastIndex(filename, path.Ext(filename))
			filename = filename[:n]

			titleMeta := ParseTitle(title)

			if tryNative && reflect.ValueOf(titleMeta).IsZero() {
				titleMeta = ParseTitle(*titleNative)
			}

			if reflect.ValueOf(titleMeta).IsZero() {
				titleMeta = ParseTitle(filename)
			}

			if !reflect.ValueOf(titleMeta).IsZero() {
				if titleMeta.Title != "" && (!hasTitleTranslated || overwrite) {
					if gallery.Translated != nil && *gallery.Translated {
						gallery.TitleTranslated = &titleMeta.Title
					} else {
						gallery.TitleNative = &titleMeta.Title
					}
				}
				if titleMeta.Released != "" && (!hasRelease || overwrite) {
					gallery.Released = &titleMeta.Released
				}
				if len(titleMeta.Artists) != 0 && titleMeta.Circle != "" && (!hasCircle || overwrite) {
					if !containsTag(currentTags, "circle", &titleMeta.Circle) {
						currentTags = append(currentTags, model.Tag{
							Namespace: "circle",
							Name:      titleMeta.Circle,
						})
					}
				}
				if len(titleMeta.Artists) != 0 && (!hasArtists || overwrite) {
					if titleMeta.Circle != "" && len(titleMeta.Artists) == 1 {
						if !containsTag(currentTags, "circle", &titleMeta.Artists[0]) {
							currentTags = append(currentTags, model.Tag{
								Namespace: "circle",
								Name:      titleMeta.Artists[0],
							})
						}
					} else {
						for _, artist := range titleMeta.Artists {
							if !containsTag(currentTags, "circle", &artist) {
								currentTags = append(currentTags, model.Tag{
									Namespace: "artist",
									Name:      artist,
								})
							}
						}
					}
				}

				// If structured, no need to set the series again.
				if library.Layout != config.Structured && titleMeta.Series != "" && (!hasSeries || overwrite) {
					gallery.Series = &titleMeta.Series
				}

				// Set as language if it's not already set and is found in the list predefined of languages.
				if titleMeta.Language != "" && (!hasLanguage || overwrite) {
					if languages[strings.ToLower(titleMeta.Language)] {
						gallery.Language = &titleMeta.Language
					} else if match, err := regexp.MatchString(`\d+`, titleMeta.Language); err == nil && match {
						exhGid, err := strconv.ParseInt(titleMeta.Language, 10, 32)
						if err == nil {
							exhGidInt32 := int32(exhGid)
							currentReference.ExhGid = &exhGidInt32
						}
					}
				}
			}

			// If the gallery is stored in a structured dir layout with no category assigned, assume it's a manga.
			if library.Layout == config.Structured && (gallery.Category == nil || *gallery.Category == "") {
				manga := "manga"
				gallery.Category = &manga
			}

			err = db.UpdateGallery(gallery, currentTags, currentReference, true)
			if err != nil {
				log.Z.Error("failed to update gallery based on its title",
					zap.String("gallery", gallery.UUID),
					zap.String("err", err.Error()))
			}
		}
	}
}

// ParseTitle parses the filename or title following the standard:
// (Release) [Circle (Artist)] Title (Series) [<usually> Language] or (Release) [Artist] Title (Series) [<usually> Language]
func ParseTitle(title string) TitleMeta {
	match := nameRegex.FindStringSubmatch(title)
	var artists []string
	if match[3] != "" {
		if strings.Contains(match[3], ", ") {
			artists = strings.Split(strings.TrimSpace(match[3]), ", ")
		} else if strings.Contains(match[3], "、") {
			artists = strings.Split(strings.TrimSpace(match[3]), "、")
		} else {
			artists = append(artists, strings.TrimSpace(match[3]))
		}
	}

	return TitleMeta{
		Released: strings.TrimSpace(match[1]),
		Circle:   strings.TrimSpace(match[2]),
		Artists:  strings.Split(strings.TrimSpace(match[3]), ", "),
		Title:    strings.TrimSpace(match[4]),
		Series:   strings.TrimSpace(match[5]),
		Language: strings.TrimSpace(match[6]),
	}
}

func containsTag(tags []model.Tag, namespace string, name *string) bool {
	for _, tag := range tags {
		if tag.Namespace == namespace && (name == nil || tag.Name == *name) {
			return true
		}
	}
	return false
}
