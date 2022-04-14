package db

import (
	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/model"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exec"
	log "github.com/sirupsen/logrus"
	"time"
)

func StorePaths(givenLibraries []config.Library) error {
	for _, library := range givenLibraries {
		libraries, err := getLibrary(library.ID, "")
		if err != nil {
			log.Error(err)
			continue
		}

		if len(libraries) == 0 {
			if err := newLibrary(library.ID, library.Path, library.Layout); err != nil {
				return err
			}
			continue
		}

		if libraries[0].Path != library.Path || libraries[0].Layout != library.Layout {
			if err := updateLibrary(library.ID, library.Path, library.Layout); err != nil {
				return err
			}
		}
	}

	return nil
}

func GetOnlyLibraries() ([]model.Library, error) {
	var libraries []model.Library
	err := database.QB().
		From("library").
		ScanStructs(&libraries)

	return libraries, err
}

type LibraryRow struct {
	ID              int32     `db:"id"`
	Path            string    `db:"path"`
	Layout          string    `db:"layout"`
	UUID            string    `db:"uuid"`
	LibraryID       int32     `db:"library_id"`
	ArchivePath     string    `db:"archive_path"`
	Title           string    `db:"title"`
	TitleNative     *string   `db:"title_native"`
	TitleTranslated *string   `db:"title_translated"`
	Category        *string   `db:"category"`
	Series          *string   `db:"series"`
	Released        *string   `db:"released"`
	Language        *string   `db:"language"`
	Translated      *bool     `db:"translated"`
	Nsfw            bool      `db:"nsfw"`
	Hidden          bool      `db:"hidden"`
	ImageCount      *int32    `db:"image_count"`
	ArchiveSize     *int32    `db:"archive_size"`
	ArchiveHash     *string   `db:"archive_hash"`
	Thumbnail       *string   `db:"thumbnail"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

func GetLibraries() ([]model.CombinedLibrary, error) {
	scanner, err := database.QB().
		From("library").
		Join(
			goqu.T("gallery"),
			goqu.On(goqu.I("gallery.library_id").Eq(goqu.I("library.id"))),
		).
		Executor().
		Scanner()

	if err != nil {
		log.Error(err)
		return nil, err
	}

	defer func(scanner exec.Scanner) {
		if err := scanner.Close(); err != nil {
			log.Error(err)
		}
	}(scanner)

	librariesMap := make(map[int32]model.CombinedLibrary)
	for scanner.Next() {
		lr := LibraryRow{}
		if err = scanner.ScanStruct(&lr); err != nil {
			log.Error(err)
			return nil, err
		}

		var gallery = model.Gallery{UUID: lr.UUID,
			Title:           lr.Title,
			TitleNative:     lr.TitleNative,
			TitleTranslated: lr.TitleTranslated,
			Category:        lr.Category,
			Series:          lr.Series,
			Released:        lr.Released,
			Language:        lr.Language,
			Translated:      lr.Translated,
			Nsfw:            lr.Nsfw,
			Hidden:          lr.Hidden,
			ImageCount:      lr.ImageCount,
			ArchiveSize:     lr.ArchiveSize,
			ArchiveHash:     lr.ArchiveHash,
			Thumbnail:       lr.Thumbnail,
			CreatedAt:       lr.CreatedAt,
			UpdatedAt:       lr.UpdatedAt,
		}

		value, ok := librariesMap[lr.ID]
		if ok {
			value.Galleries = append(value.Galleries, gallery)
			librariesMap[lr.ID] = value
		} else {
			librariesMap[lr.ID] = model.CombinedLibrary{
				Library: model.Library{
					ID:     lr.ID,
					Path:   lr.Path,
					Layout: lr.Layout,
				},
				Galleries: []model.Gallery{gallery},
			}
		}
	}

	librariesSlice := make([]model.CombinedLibrary, 0, len(librariesMap))
	for _, val := range librariesMap {
		librariesSlice = append(librariesSlice, val)
	}

	return librariesSlice, nil
}

// getLibrary returns the library from the database based on the ID or path.
func getLibrary(id int32, path string) ([]model.Library, error) {
	stmt := database.QB().From("library").Prepared(true)

	if path == "" {
		stmt = stmt.Where(goqu.Ex{
			"id": id,
		})
	} else {
		stmt = stmt.Where(goqu.ExOr{
			"id":   id,
			"path": path,
		})
	}

	var libraries []model.Library
	err := stmt.ScanStructs(&libraries)

	return libraries, err
}

// newLibrary creates a new library to the database.
func newLibrary(id int32, path string, layout string) error {
	_, err := database.QB().
		Insert("library").
		Prepared(true).
		Rows(goqu.Record{
			"id":     id,
			"path":   path,
			"layout": layout,
		}).
		Executor().
		Exec()

	if err != nil {
		err = updateLibrary(id, path, layout)
	}

	return err
}

// updateLibrary updates the library in the database.
func updateLibrary(id int32, path string, layout string) error {
	_, err := database.QB().
		Update("library").
		Prepared(true).
		Set(goqu.Record{
			"path":   path,
			"layout": layout,
		}).
		Where(goqu.Ex{
			"id": id,
		}).
		Executor().
		Exec()

	return err
}
