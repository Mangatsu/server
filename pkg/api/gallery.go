package api

import (
	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/library"
	"github.com/Mangatsu/server/pkg/types/model"
	"github.com/gorilla/mux"
	"net/http"
)

type MetadataResult struct {
	Hidden      bool   `json:",omitempty"`
	ArchivePath string `json:",omitempty"`
	LibraryID   string `json:",omitempty"`
	model.Gallery

	Tags map[string][]string

	Reference struct {
		ExhToken *string
		ExhGid   *int32
		Urls     *string
	} `alias:"reference.*"`

	GalleryPref *struct {
		FavoriteGroup *string
		Progress      int32
		UpdatedAt     string
	} `alias:"gallery_pref.*"`

	Library model.Library `json:"-"`
}

type GalleryResult struct {
	Meta  MetadataResult
	Files []string
	Count int
}

type GenericStringResult struct {
	Data  []string
	Count int
}

// returnGalleries returns galleries as JSON.
func returnGalleries(w http.ResponseWriter, r *http.Request) {
	access, userUUID := hasAccess(w, r, db.NoRole)
	if !access {
		return
	}

	queryParams := parseQueryParams(r)
	galleries, err := db.GetGalleries(queryParams, true, userUUID)
	if handleResult(w, galleries, err, false) {
		return
	}

	var galleriesResult []MetadataResult
	count := len(galleries)
	if count > 0 {
		for _, gallery := range galleries {
			galleriesResult = append(galleriesResult, convertMetadata(gallery))
		}
	}

	resultToJSON(w, struct {
		Data  []MetadataResult
		Count int
	}{
		Data:  galleriesResult,
		Count: count,
	})
}

// returnGallery returns one gallery as JSON.
func returnGallery(w http.ResponseWriter, r *http.Request) {
	access, userUUID := hasAccess(w, r, db.NoRole)
	if !access {
		return
	}

	params := mux.Vars(r)
	galleryUUID := params["uuid"]

	galleries, err := db.GetGallery(&galleryUUID, userUUID)
	if handleResult(w, galleries, err, false) {
		return
	}

	gallery := convertMetadata(galleries[0])
	files, count := library.ReadGallery(
		config.BuildLibraryPath(gallery.Library.Path, gallery.ArchivePath),
		gallery.UUID,
	)

	resultToJSON(w, GalleryResult{
		Meta:  gallery,
		Files: files,
		Count: count,
	})
}

// returnRandomGallery returns one random gallery as JSON in the same way as returnGallery.
func returnRandomGallery(w http.ResponseWriter, r *http.Request) {
	access, userUUID := hasAccess(w, r, db.NoRole)
	if !access {
		return
	}

	galleries, err := db.GetGallery(nil, userUUID)
	if handleResult(w, galleries, err, false) {
		return
	}

	gallery := convertMetadata(galleries[0])
	files, count := library.ReadGallery(
		config.BuildLibraryPath(gallery.Library.Path, gallery.ArchivePath),
		gallery.UUID,
	)

	resultToJSON(w, GalleryResult{
		Meta:  gallery,
		Files: files,
		Count: count,
	})
}

// returnTags returns all tags as JSON.
func returnTags(w http.ResponseWriter, r *http.Request) {
	if access, _ := hasAccess(w, r, db.NoRole); !access {
		return
	}

	tags, err := db.GetTags()
	if handleResult(w, tags, err, true) {
		return
	}

	resultToJSON(w, tags)
}

// returnCategories returns all public categories as JSON.
func returnCategories(w http.ResponseWriter, r *http.Request) {
	if access, _ := hasAccess(w, r, db.NoRole); !access {
		return
	}

	categories, err := db.GetCategories()
	if handleResult(w, categories, err, true) {
		return
	}

	resultToJSON(w, GenericStringResult{
		Data:  categories,
		Count: len(categories),
	})
}

// returnSeries returns all series as JSON.
func returnSeries(w http.ResponseWriter, r *http.Request) {
	if access, _ := hasAccess(w, r, db.NoRole); !access {
		return
	}

	series, err := db.GetSeries()
	if handleResult(w, series, err, true) {
		return
	}

	resultToJSON(w, GenericStringResult{
		Data:  series,
		Count: len(series),
	})
}
