package api

import (
	"encoding/json"
	"github.com/Luukuton/Mangatsu/internal/config"
	"github.com/Luukuton/Mangatsu/pkg/db"
	"github.com/Luukuton/Mangatsu/pkg/library"
	"github.com/Luukuton/Mangatsu/pkg/types/model"
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

type GalleriesResult struct {
	Data  []MetadataResult `json:"Data"`
	Count int32            `json:"Count"`
}

type GalleryResult struct {
	Meta  MetadataResult `json:"Meta"`
	Files []string       `json:"Files"`
	Count int32          `json:"Count"`
}

type GenericStringResult struct {
	Data  []string `json:"Data"`
	Count int32    `json:"Count"`
}

// returnGalleries returns galleries as JSON.
func returnGalleries(w http.ResponseWriter, r *http.Request) {
	access, userUUID := hasAccess(w, r, db.NoRole)
	if !access {
		return
	}

	queryParams := parseQueryParams(r)
	galleries := db.GetGalleries(queryParams, true, userUUID)

	var galleriesResult []MetadataResult
	if len(galleries) == 0 {
		galleriesResult = []MetadataResult{}
	} else {
		for _, gallery := range galleries {
			galleriesResult = append(galleriesResult, convertMetadata(gallery))
		}
	}

	result := GalleriesResult{
		Data:  galleriesResult,
		Count: int32(len(galleries)),
	}

	//w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	//w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	err := json.NewEncoder(w).Encode(result)
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "")
		return
	}
}

// returnGallery returns one gallery as JSON.
func returnGallery(w http.ResponseWriter, r *http.Request) {
	access, userUUID := hasAccess(w, r, db.NoRole)
	if !access {
		return
	}

	params := mux.Vars(r)
	galleryUUID := params["uuid"]

	galleries := db.GetGallery(&galleryUUID, userUUID)
	if len(galleries) == 0 {
		errorHandler(w, http.StatusNotFound, "")
		return
	}

	gallery := convertMetadata(galleries[0])
	files, count := library.ReadGallery(
		config.BuildLibraryPath(gallery.Library.Path, gallery.ArchivePath),
		gallery.UUID,
	)
	result := GalleryResult{
		Meta:  gallery,
		Files: files,
		Count: count,
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	err := json.NewEncoder(w).Encode(result)
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "")
		return
	}
}

// returnRandomGallery returns one random gallery as JSON in the same way as returnGallery.
func returnRandomGallery(w http.ResponseWriter, r *http.Request) {
	access, userUUID := hasAccess(w, r, db.NoRole)
	if !access {
		return
	}

	galleries := db.GetGallery(nil, userUUID)
	if len(galleries) == 0 {
		galleries = []db.CombinedMetadata{}
	}

	gallery := convertMetadata(galleries[0])
	files, count := library.ReadGallery(
		config.BuildLibraryPath(gallery.Library.Path, gallery.ArchivePath),
		gallery.UUID,
	)
	result := GalleryResult{
		Meta:  gallery,
		Files: files,
		Count: count,
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	err := json.NewEncoder(w).Encode(result)
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "")
		return
	}
}

// returnTag Returns one tag with galleries associated with it as JSON.
func returnTag(w http.ResponseWriter, r *http.Request) {
	if access, _ := hasAccess(w, r, db.NoRole); !access {
		return
	}

	params := mux.Vars(r)
	tags := db.GetTag(params["namespace"], params["name"])
	if tags == nil || len(tags) == 0 {
		errorHandler(w, http.StatusNotFound, "")
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	err := json.NewEncoder(w).Encode(tags)
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "")
		return
	}
}

// returnTags returns all tags as JSON.
func returnTags(w http.ResponseWriter, r *http.Request) {
	if access, _ := hasAccess(w, r, db.NoRole); !access {
		return
	}

	tags := db.GetTags()
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	err := json.NewEncoder(w).Encode(tags)
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "")
		return
	}
}

// returnCategories returns all public categories as JSON.
func returnCategories(w http.ResponseWriter, r *http.Request) {
	if access, _ := hasAccess(w, r, db.NoRole); !access {
		return
	}

	categories := db.GetCategories()
	response := GenericStringResult{
		Data:  categories,
		Count: int32(len(categories)),
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "")
		return
	}
}

// returnSeries returns all public categories as JSON.
func returnSeries(w http.ResponseWriter, r *http.Request) {
	if access, _ := hasAccess(w, r, db.NoRole); !access {
		return
	}

	series := db.GetSeries()
	response := GenericStringResult{
		Data:  series,
		Count: int32(len(series)),
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, "")
		return
	}
}
