package db

import (
	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/types/model"
	"github.com/doug-martin/goqu/v9"
	log "github.com/sirupsen/logrus"
)

// FIXME: it won't work yet!! CombinedLibrary has to be a plain struct,
// something like this:
//
// 	type User struct {
// 		FirstName string `db:"first_name"`
// 		LastName  string `db:"last_name"`
//	}

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
	var libraries []model.Library
	err := database.QB().
		From("library").
		ScanStructs(&libraries)

	return libraries, err
}

func GetLibraries() ([]CombinedLibrary, error) {
	var libraries []CombinedLibrary
	err := database.QB().
		From("library").
		LeftJoin(
			goqu.T("gallery"),
			goqu.On(goqu.Ex{
				"gallery.id": goqu.I("library.id"),
			}),
		).
		ScanStructs(&libraries)

	return libraries, err
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
			"id": id,
			"path": path,
		})
	}

	var libraries []model.Library
	err := stmt.ScanStructs(&libraries)

	return libraries, err
}

// newLibrary creates a new library to the database.
func newLibrary(id int32, path string, layout string) error {
	// does not handle the scenario when the ID exists in db
	_, err := database.QB().
		Insert("library").
		Prepared(true).
		Rows(goqu.Record{
			"id": id,
			"path": path,
			"layout": layout,
		}).
		Executor().
		Exec()

	return err
}

// updateLibrary updates the library in the database.
func updateLibrary(id int32, path string, layout string) error {
	_, err := database.QB().
		Update("library").
		Prepared(true).
		Set(goqu.Record{
			"path": path,
			"layout": layout,
		}).
		Where(goqu.Ex{
			"id": id,
		}).
		Executor().
		Exec()

	return err
}
