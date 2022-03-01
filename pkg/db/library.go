package db

import (
	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/types/model"
	. "github.com/Mangatsu/server/pkg/types/table"
	. "github.com/go-jet/jet/v2/sqlite"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

type CombinedLibrary struct {
	model.Library
	Galleries []model.Gallery
}

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
	stmt := SELECT(Library.AllColumns).FROM(Library.Table)
	var libraries []model.Library

	err := stmt.Query(db(), &libraries)
	return libraries, err
}

func GetLibraries() ([]CombinedLibrary, error) {
	stmt := SELECT(Library.AllColumns, Gallery.AllColumns).
		FROM(Library.LEFT_JOIN(Gallery, Gallery.LibraryID.EQ(Library.ID)))
	var libraries []CombinedLibrary

	err := stmt.Query(db(), &libraries)
	return libraries, err
}

// getLibrary returns the library from the database based on the ID or path.
func getLibrary(id int32, path string) ([]model.Library, error) {
	stmt := SELECT(
		Library.AllColumns,
	).FROM(
		Library.Table,
	)

	if path == "" {
		stmt = stmt.WHERE(Library.ID.EQ(Int32(id)))
	} else {
		stmt = stmt.WHERE(Library.ID.EQ(Int32(id)).OR(Library.Path.EQ(String(path))))
	}

	var libraries []model.Library
	err := stmt.Query(db(), &libraries)
	return libraries, err
}

// newLibrary creates a new library to the database.
func newLibrary(id int32, path string, layout string) error {
	stmt := Library.INSERT(Library.ID, Library.Path, Library.Layout).VALUES(id, path, layout).
		ON_CONFLICT(Library.ID).
		DO_UPDATE(SET(Library.Path.SET(String(path)), Library.Layout.SET(String(layout))))

	_, err := stmt.Exec(db())
	if err != nil {
		log.Error(err)
	}

	return err
}

// updateLibrary updates the library in the database.
func updateLibrary(id int32, path string, layout string) error {
	stmt := Library.UPDATE(Library.Path, Library.Layout).SET(path, layout).WHERE(Library.ID.EQ(Int32(id)))
	_, err := stmt.Exec(db())
	return err
}
