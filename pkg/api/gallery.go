package api

import (
	"encoding/json"
	"fmt"
	"github.com/Mangatsu/server/pkg/utils"
	"net/http"

	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/cache"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/log"
	"github.com/Mangatsu/server/pkg/types/model"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type MetadataResult struct {
	Hidden          bool    `json:",omitempty"`
	ArchivePath     string  `json:",omitempty"`
	LibraryID       string  `json:",omitempty"`
	SubGalleryCount *uint64 `json:",omitempty"`
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
	galleries, totalCount, err := db.GetGalleries(queryParams, true, userUUID)
	if handleResult(w, galleries, err, true, r.RequestURI) {
		return
	}

	var galleriesResult []MetadataResult
	count := len(galleries)
	if count > 0 {
		for _, gallery := range galleries {
			galleriesResult = append(galleriesResult, convertMetadata(gallery))
		}
	}

	grouped := queryParams.Grouped == "true"
	groupedResult := utils.NewOrderedMap()
	if grouped {
		for _, gallery := range galleriesResult {
			if gallery.Series != nil && *gallery.Series != "" {
				subGalleriesCount, err := db.GetGalleryCount(db.Filters{Series: *gallery.Series}, true, userUUID)
				if err != nil {
					log.Z.Debug("failed getting sub gallery count",
						zap.Stringp("name", gallery.Series),
						zap.String("err", err.Error()))
					continue
				}

				gallery.SubGalleryCount = &subGalleriesCount
				groupedResult.Set(*gallery.Series, gallery)
			} else {
				subGalleriesCount := uint64(1)
				gallery.SubGalleryCount = &subGalleriesCount
				groupedResult.Set(gallery.UUID, gallery)
			}
		}
	}

	if grouped {
		resultToJSON(w, struct {
			Data       *utils.OrderedMap
			Count      int
			TotalCount uint64
		}{
			Data:       groupedResult,
			Count:      count,
			TotalCount: totalCount,
		}, r.RequestURI)
		return
	}

	resultToJSON(w, struct {
		Data       []MetadataResult
		Count      int
		TotalCount uint64
	}{
		Data:       galleriesResult,
		Count:      count,
		TotalCount: totalCount,
	}, r.RequestURI)
}

// returnGalleryCount returns the amount of galleries.
func returnGalleryCount(w http.ResponseWriter, r *http.Request) {
	access, userUUID := hasAccess(w, r, db.NoRole)
	if !access {
		return
	}

	queryParams := parseQueryParams(r)
	count, err := db.GetGalleryCount(queryParams, true, userUUID)
	if handleResult(w, count, err, false, r.URL.Path) {
		return
	}

	resultToJSON(w, struct{ Count uint64 }{Count: count}, r.URL.Path)
}

// returnGallery returns one gallery as JSON.
func returnGallery(w http.ResponseWriter, r *http.Request) {
	access, userUUID := hasAccess(w, r, db.NoRole)
	if !access {
		return
	}

	params := mux.Vars(r)
	galleryUUID := params["uuid"]

	gallery, err := db.GetGallery(&galleryUUID, userUUID)
	if handleResult(w, gallery, err, false, r.RequestURI) {
		return
	}

	galleryWithMeta := convertMetadata(gallery)
	if r.URL.Query().Get("meta") == "true" {
		resultToJSON(w, GalleryResult{
			Meta:  galleryWithMeta,
			Files: nil,
			Count: 0,
		}, r.RequestURI)
		return
	}

	galleryPath := config.BuildLibraryPath(galleryWithMeta.Library.Path, galleryWithMeta.ArchivePath)
	files, count := cache.Read(galleryPath, galleryWithMeta.UUID)
	resultToJSON(w, GalleryResult{
		Meta:  galleryWithMeta,
		Files: files,
		Count: count,
	}, r.RequestURI)
}

// returnRandomGallery returns one random gallery as JSON in the same way as returnGallery.
func returnRandomGallery(w http.ResponseWriter, r *http.Request) {
	access, userUUID := hasAccess(w, r, db.NoRole)
	if !access {
		return
	}

	gallery, err := db.GetGallery(nil, userUUID)
	if handleResult(w, gallery, err, false, r.URL.Path) {
		return
	}

	galleryWithMeta := convertMetadata(gallery)
	galleryPath := config.BuildLibraryPath(galleryWithMeta.Library.Path, galleryWithMeta.ArchivePath)
	files, count := cache.Read(galleryPath, galleryWithMeta.UUID)

	resultToJSON(w, GalleryResult{
		Meta:  galleryWithMeta,
		Files: files,
		Count: count,
	}, r.URL.Path)
}

// returnTags returns all tags as JSON.
func returnTags(w http.ResponseWriter, r *http.Request) {
	if access, _ := hasAccess(w, r, db.NoRole); !access {
		return
	}

	tags, _, err := db.GetTags("", true)
	if handleResult(w, tags, err, true, r.RequestURI) {
		return
	}

	resultToJSON(w, tags, r.RequestURI)
}

// returnCategories returns all public categories as JSON.
func returnCategories(w http.ResponseWriter, r *http.Request) {
	if access, _ := hasAccess(w, r, db.NoRole); !access {
		return
	}

	categories, err := db.GetCategories()
	if handleResult(w, categories, err, true, r.RequestURI) {
		return
	}

	resultToJSON(w, GenericStringResult{
		Data:  categories,
		Count: len(categories),
	}, r.RequestURI)
}

// returnSeries returns all series as JSON.
func returnSeries(w http.ResponseWriter, r *http.Request) {
	if access, _ := hasAccess(w, r, db.NoRole); !access {
		return
	}

	series, err := db.GetSeries()
	if handleResult(w, series, err, true, r.RequestURI) {
		return
	}

	resultToJSON(w, GenericStringResult{
		Data:  series,
		Count: len(series),
	}, r.RequestURI)
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
		errorHandler(w, http.StatusBadRequest, "gallery uuid is required", r.URL.Path)
		return
	}

	formData := &UpdateGalleryForm{}
	if err := json.NewDecoder(r.Body).Decode(formData); err != nil {
		errorHandler(w, http.StatusBadRequest, err.Error(), r.URL.Path)
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
		errorHandler(w, http.StatusInternalServerError, err.Error(), r.URL.Path)
		return
	}
	fmt.Fprintf(w, `{ "message": "gallery updated" }`)
}
