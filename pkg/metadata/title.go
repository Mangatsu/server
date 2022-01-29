package metadata

import (
	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/types/model"
	log "github.com/sirupsen/logrus"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

var nameRegex = regexp.MustCompile(
	"(?i)(?:\\(([^([]+)\\))?\\s*(?:\\[([^()[\\]]+)(?:\\(([^()[\\]]+)\\))?])?\\s*([^([]+)\\s*(?:\\(([^([)]+)\\))?\\s*(?:\\[(?:DL版?|digital)])?\\s*(?:\\[([^\\]]+)])?",
)

type TitleMeta struct {
	Released string
	Circle   string
	Artists  string
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
		log.Error("Libraries could not be retrieved to parse titles: ", err)
		return
	}

	for _, library := range libraries {
		for _, gallery := range library.Galleries {
			hasTitleShort := gallery.TitleShort != nil
			hasRelease := gallery.Released != nil
			hasCircle := gallery.Circle != nil
			hasArtists := gallery.Artists != nil
			hasSeries := gallery.Series != nil
			hasLanguage := gallery.Language != nil

			if !overwrite && hasRelease && hasCircle && hasArtists && hasSeries && hasLanguage {
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
				if titleMeta.Title != "" && (!hasTitleShort || overwrite) {
					gallery.TitleShort = &titleMeta.Title
				}
				if titleMeta.Released != "" && (!hasRelease || overwrite) {
					gallery.Released = &titleMeta.Released
				}
				if titleMeta.Artists != "" && titleMeta.Circle != "" && (!hasCircle || overwrite) {
					gallery.Circle = &titleMeta.Circle
				}
				if titleMeta.Artists != "" && (!hasArtists || overwrite) {
					if titleMeta.Circle != "" {
						gallery.Artists = &titleMeta.Circle
					} else {
						gallery.Artists = &titleMeta.Artists
					}
				}
				if titleMeta.Series != "" && (!hasSeries || overwrite) {
					gallery.Series = &titleMeta.Series
				}
				if titleMeta.Language != "" && (!hasLanguage || overwrite) {
					gallery.Language = &titleMeta.Language
				}
			}

			// If the gallery is stored in a structured dir layout with no category assigned, assume it's a manga.
			if library.Layout == config.Structured && (gallery.Category == nil || *gallery.Category == "") {
				manga := "manga"
				gallery.Category = &manga
			}

			// If structured, set the Series to the first dir name.
			if library.Layout == config.Structured {
				dirs := strings.SplitN(gallery.ArchivePath, "/", 2)
				gallery.Series = &dirs[0]
			}

			err := db.UpdateGallery(gallery, nil, model.Reference{})
			if err != nil {
				log.Errorf("Error updating gallery %s based on its title: %s", gallery.UUID, err)
			}
		}
	}
}

// ParseTitle parses the filename or title following the standard:
// (Release) [Circle (Artist)] Title (Series) [Language] or (Release) [Artist] Title (Series) [Language]
func ParseTitle(title string) TitleMeta {
	match := nameRegex.FindStringSubmatch(title)
	return TitleMeta{
		Released: strings.TrimSpace(match[1]),
		Circle:   strings.TrimSpace(match[2]),
		Artists:  strings.TrimSpace(match[3]),
		Title:    strings.TrimSpace(match[4]),
		Series:   strings.TrimSpace(match[5]),
		Language: strings.TrimSpace(match[6]),
	}
}
