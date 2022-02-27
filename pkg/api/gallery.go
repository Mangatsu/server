package api

import (
	"encoding/json"
	"fmt"
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

type UpdateGalleryForm struct {
	Title           string
	TitleNative     string
	TitleTranslated string
	Released        string
	Series          string
	Category        string
	Language        string
	Translated      bool
	Nsfw            bool
	Hidden          bool
	ExhToken        string
	ExhGid          int32
	AnilistID       int32
	Urls            string
	Tags            map[string][]string
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

	tags, _, err := db.GetTags(nil, true)
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

// updateGallery updates a gallery and its reference and tags.
// If tags field is specified and empty, all references to this gallery's tags will be removed.
// If tags is not specified, no changes to tags will be made.
func updateGallery(w http.ResponseWriter, r *http.Request) {
	access, _ := hasAccess(w, r, db.Admin)
	if !access {
		return
	}

	params := mux.Vars(r)
	galleryUUID := params["uuid"]
	if galleryUUID == "" {
		errorHandler(w, http.StatusBadRequest, "")
		return
	}

	formData := &UpdateGalleryForm{}
	if err := json.NewDecoder(r.Body).Decode(formData); err != nil {
		errorHandler(w, http.StatusBadRequest, "")
		return
	}

	newGallery := model.Gallery{
		UUID:            galleryUUID,
		Title:           formData.Title,
		TitleNative:     &formData.TitleNative,
		TitleTranslated: &formData.TitleTranslated,
		Released:        &formData.Released,
		Series:          &formData.Series,
		Category:        &formData.Category,
		Language:        &formData.Language,
		Translated:      &formData.Translated,
		Nsfw:            formData.Nsfw,
		Hidden:          formData.Hidden,
	}

	newReference := model.Reference{
		GalleryUUID: galleryUUID,
		Urls:        &formData.Urls,
		ExhToken:    &formData.ExhToken,
		ExhGid:      &formData.ExhGid,
		AnilistID:   &formData.AnilistID,
	}

	var tags []model.Tag
	if formData.Tags != nil {
		tags = []model.Tag{}
		for namespace, names := range formData.Tags {
			for _, name := range names {
				tag := model.Tag{
					Namespace: namespace,
					Name:      name,
				}
				tags = append(tags, tag)
			}
		}
	}

	if err := db.UpdateGallery(newGallery, tags, newReference, false); err != nil {
		errorHandler(w, http.StatusInternalServerError, "")
		return
	}
	fmt.Fprintf(w, `{ "message": "gallery updated" }`)
}
