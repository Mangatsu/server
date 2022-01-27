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

func StorePaths(libraries []config.Library) {
	for _, library := range libraries {
		libraries := getLibrary(library.ID, "")
		if len(libraries) == 0 {
			err := newLibrary(library.ID, library.Path, library.Layout)
			if err != nil {
				log.Fatal("Error saving library to db: ", err)
			}
			continue
		}

		if libraries[0].Path != library.Path || libraries[0].Layout != library.Layout {
			err := updateLibrary(library.ID, library.Path, library.Layout)
			if err != nil {
				log.Fatal("Error updating library in db: ", err)
			}
		}
	}
}

func GetOnlyLibraries() []model.Library {
	stmt := SELECT(Library.AllColumns).FROM(Library.Table)
	var libraries []model.Library
	err := stmt.Query(db(), &libraries)
	if err != nil {
		log.Error(err)
		return nil
	}

	return libraries
}

func GetLibraries() []CombinedLibrary {
	stmt := SELECT(Library.AllColumns, Gallery.AllColumns).
		FROM(Library.INNER_JOIN(Gallery, Gallery.LibraryID.EQ(Library.ID)))

	var libraries []CombinedLibrary
	err := stmt.Query(db(), &libraries)
	if err != nil {
		log.Error(err)
		return nil
	}

	return libraries
}

// getLibrary returns the library from the database based on the ID or path.
func getLibrary(id int32, path string) []model.Library {
	selectStmt := SELECT(
		Library.AllColumns,
	).FROM(
		Library.Table,
	)

	if path == "" {
		selectStmt = selectStmt.WHERE(Library.ID.EQ(Int32(id)))
	} else {
		selectStmt = selectStmt.WHERE(Library.ID.EQ(Int32(id)).OR(Library.Path.EQ(String(path))))
	}

	var libraries []model.Library
	err := selectStmt.Query(db(), &libraries)
	if err != nil {
		log.Error(err)
		return nil
	}

	return libraries
}

// newLibrary creates a new library to the database.
func newLibrary(id int32, path string, layout string) error {
	insertStmt := Library.INSERT(Library.ID, Library.Path, Library.Layout).VALUES(id, path, layout).
		ON_CONFLICT(Library.ID).
		DO_UPDATE(SET(Library.Path.SET(String(path)), Library.Layout.SET(String(layout))))

	_, err := insertStmt.Exec(db())
	if err != nil {
		log.Error(err)
	}

	return err
}

// updateLibrary updates the library in the database.
func updateLibrary(id int32, path string, layout string) error {
	updateLibrary := Library.UPDATE(Library.Path, Library.Layout).SET(path, layout).WHERE(Library.ID.EQ(Int32(id)))

	_, err := updateLibrary.Exec(db())
	if err != nil {
		log.Error(err)
	}

	return err
}
