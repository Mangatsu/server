package metadata

import (
	"encoding/json"
	"fmt"
	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/library"
	"github.com/Mangatsu/server/pkg/types/model"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"io/ioutil"
	"math"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type Tags map[string][]string

type XMetadata struct {
	GalleryInfo struct {
		Title         *string `json:"title"`
		TitleOriginal *string `json:"title_original"`
		Link          *string `json:"link"`
		Category      *string `json:"category"`
		Tags          Tags    `json:"tags"`
		Language      *string `json:"language"`
		Translated    *bool   `json:"translated"`
		UploadDate    *[]int  `json:"upload_date"`
		Source        *struct {
			Site  *string `json:"site"`
			Gid   *int32  `json:"gid"`
			Token *string `json:"token"`
		} `json:"source"`
	} `json:"gallery_info"`
	GalleryInfoFull *struct {
		Gallery struct {
			Gid   *int32  `json:"gid"`
			Token *string `json:"token"`
		} `json:"gallery"`
		Title               *string `json:"title"`
		TitleOriginal       *string `json:"title_original"`
		DateUploaded        *int64  `json:"date_uploaded"`
		Category            *string `json:"category"`
		Uploader            *string `json:"uploader"`
		ImageCount          *int32  `json:"image_count"`
		ImagesResized       *bool   `json:"images_resized"`
		TotalFileSizeApprox *int32  `json:"total_file_size_approx"`
		Language            *string `json:"language"`
		Translated          *bool   `json:"translated"`
		Tags                Tags    `json:"tags"`
		TagsHaveNamespace   *bool   `json:"tags_have_namespace"`
		Source              *string `json:"source"`
		SourceSite          *string `json:"source_site"`
	} `json:"gallery_info_full"`
}

var metaExtensions = regexp.MustCompile(`\.json$`)

type NoMatchPaths struct {
	libraryPath string
	fullPath    string
}

var archivesNoMatch []NoMatchPaths

// unmarshalExhJSON parses ExH JSON bytes into XMetadata.
func unmarshalExhJSON(byteValue []byte) (XMetadata, error) {
	var gallery XMetadata
	err := json.Unmarshal(byteValue, &gallery)
	if err != nil {
		log.Error("Error in unmarshalling x metadata: ", err)
		return XMetadata{}, err
	}

	return gallery, err
}

// convertExh converts ExH model to gallery, tags and other models.
func convertExh(
	exhGallery XMetadata,
	archivePath string,
	metaPath string,
	internal bool,
) (model.Gallery, []model.Tag, model.Reference) {
	title := archivePath
	if exhGallery.GalleryInfo.Title == nil {
		title = path.Base(archivePath)
		n := strings.LastIndex(title, path.Ext(title))
		title = title[:n]
	} else {
		title = *exhGallery.GalleryInfo.Title
	}

	newGallery := model.Gallery{
		Title:       title,
		TitleNative: exhGallery.GalleryInfo.TitleOriginal,
		Category:    exhGallery.GalleryInfo.Category,
		Language:    exhGallery.GalleryInfo.Language,
		Translated:  exhGallery.GalleryInfo.Translated,
		ImageCount:  exhGallery.GalleryInfoFull.ImageCount,
		ArchiveSize: exhGallery.GalleryInfoFull.TotalFileSizeApprox,
		ArchivePath: archivePath,
		Nsfw:        *exhGallery.GalleryInfo.Category != "non-h",
	}

	var tags []model.Tag
	for namespace, names := range exhGallery.GalleryInfo.Tags {
		for _, name := range names {
			tags = append(tags, model.Tag{Namespace: namespace, Name: name})
		}
	}

	exh := model.Reference{
		MetaPath:     &metaPath,
		MetaInternal: internal,
		ExhGid:       exhGallery.GalleryInfo.Source.Gid,
		ExhToken:     exhGallery.GalleryInfo.Source.Token,
		Urls:         nil,
	}

	return newGallery, tags, exh
}

//func needsUpdate(archivePath string) bool {
//	stat, err := os.Stat(archivePath)
//	if err != nil {
//		log.Error("Error in getting archive stats: ", err)
//		return false
//	}
//
//	modTime := stat.ModTime()
//	//size := stat.Size()
//	needsUpdate, _ := db.NeedsUpdate(archivePath, modTime)
//
//	return needsUpdate
//}

// matchInternalMeta tries to find the metadata file either in the archive.
func matchInternalMeta(fullArchivePath string) ([]byte, string) {
	metaData, metaSource := library.ReadArchiveInternalMeta(fullArchivePath)
	return metaData, metaSource
}

// matchExternalMeta tries to find the metadata file besides it (exact match).
func matchExternalMeta(fullArchivePath string, libraryPath string) ([]byte, string) {
	externalJSON := library.ArchiveExtensions.ReplaceAllString(fullArchivePath, ".json")

	if !library.PathExists(externalJSON) {
		archivesNoMatch = append(archivesNoMatch, NoMatchPaths{libraryPath: libraryPath, fullPath: fullArchivePath})
		return nil, ""
	}

	metaData, err := library.ReadJSON(externalJSON)
	if err != nil {
		log.Debug("Couldn't read external metadata: ", err)
		return nil, ""
	}

	return metaData, config.RelativePath(libraryPath, externalJSON)
}

type FuzzyResult struct {
	MetaTitleMatch     bool
	Similarity         float64
	MatchedArchivePath string
	RelativeMetaPath   string
}

// fuzzyMatchExternalMeta tries to find the metadata file besides it (fuzzy match).
func fuzzyMatchExternalMeta(archivePath string, libraryPath string, f fs.FileInfo) (FuzzyResult, XMetadata) {
	fuzzyResult := FuzzyResult{
		MetaTitleMatch:     false,
		Similarity:         0.0,
		MatchedArchivePath: "",
	}

	if f.IsDir() || !metaExtensions.MatchString(f.Name()) {
		return fuzzyResult, XMetadata{}
	}

	archivePath = filepath.ToSlash(archivePath)
	onlyDir := filepath.Dir(archivePath)

	metaData, err := library.ReadJSON(config.BuiltPath(onlyDir, f.Name()))
	if err != nil {
		log.Error(err)
		return fuzzyResult, XMetadata{}
	}

	exhGallery, err := unmarshalExhJSON(metaData)
	if err != nil {
		log.Debug("Couldn't unmarshal exhGallery: ", err)
		return fuzzyResult, XMetadata{}
	}

	relativeMetaPath := config.RelativePath(libraryPath, onlyDir+"/"+f.Name())
	// Skip if the JSON metadata has already been used by another archive.
	if db.MetaPathFound(relativeMetaPath, libraryPath) {
		return FuzzyResult{}, XMetadata{}
	}

	fuzzyResult.RelativeMetaPath = relativeMetaPath
	relativeArchivePath := config.RelativePath(libraryPath, archivePath)
	metaNoExt := metaExtensions.ReplaceAllString(f.Name(), "")
	archiveNoExt := library.ArchiveExtensions.ReplaceAllString(path.Base(relativeArchivePath), "")

	archiveSimilarity := library.Similarity(archiveNoExt, metaNoExt)
	titleSimilarity := library.Similarity(archiveNoExt, *exhGallery.GalleryInfo.Title)
	titleNativeSimilarity := library.Similarity(archiveNoExt, *exhGallery.GalleryInfo.TitleOriginal)
	fuzzyResult.MetaTitleMatch = *exhGallery.GalleryInfo.Title == archiveNoExt || *exhGallery.GalleryInfo.TitleOriginal == archiveNoExt

	if fuzzyResult.MetaTitleMatch {
		fuzzyResult.MatchedArchivePath = relativeArchivePath
		return fuzzyResult, exhGallery
	}

	if archiveSimilarity > fuzzyResult.Similarity {
		fuzzyResult.MatchedArchivePath = relativeArchivePath
		fuzzyResult.Similarity = archiveSimilarity
	}

	if titleSimilarity > fuzzyResult.Similarity {
		fuzzyResult.MatchedArchivePath = relativeArchivePath
		fuzzyResult.Similarity = titleSimilarity
	}

	if titleNativeSimilarity > fuzzyResult.Similarity {
		fuzzyResult.MatchedArchivePath = relativeArchivePath
		fuzzyResult.Similarity = titleNativeSimilarity
	}

	return fuzzyResult, exhGallery
}

// ParseX parses x JSON files (x: https://github.com/dnsev-h/x).
func ParseX() {
	libraries, err := db.GetLibraries()
	if err != nil {
		log.Error("Libraries could not be retrieved to parse x JSON files: ", err)
		return
	}
	for _, galleryLibrary := range libraries {
		for _, gallery := range galleryLibrary.Galleries {
			fullPath := config.BuildLibraryPath(galleryLibrary.Path, gallery.ArchivePath)
			//if !needsUpdate(fullPath) {
			//	continue
			//}

			var metaData []byte
			var metaPath string
			internal := false

			metaData, metaPath = matchInternalMeta(fullPath)
			if metaData != nil {
				internal = true
			}

			if metaData == nil {
				metaData, metaPath = matchExternalMeta(fullPath, galleryLibrary.Path)
			}

			if metaData != nil {
				exhGallery, err := unmarshalExhJSON(metaData)
				if err != nil {
					log.Debug("Couldn't unmarshal JSON: ", err)
					continue
				}

				newGallery, tags, external := convertExh(exhGallery, gallery.ArchivePath, metaPath, internal)
				if err = db.UpdateGallery(newGallery, tags, external); err != nil {
					log.Debugf("Couldn't tag gallery: %s. Message: %s", gallery.ArchivePath, err)
					continue
				}
			}
		}
	}

	// Fuzzy parsing for leftover archives.
	for _, noMatch := range archivesNoMatch {
		onlyDir := filepath.Dir(noMatch.fullPath)
		files, err := ioutil.ReadDir(onlyDir)
		if err != nil {
			log.Debug("Couldn't read dir while fuzzy matching: ", err)
		}

		for _, f := range files {
			r, exhGallery := fuzzyMatchExternalMeta(noMatch.fullPath, noMatch.libraryPath, f)

			if r.MatchedArchivePath != "" && r.MetaTitleMatch || r.Similarity > 0.70 {
				gallery, tags, external := convertExh(exhGallery, r.MatchedArchivePath, r.RelativeMetaPath, false)

				if !r.MetaTitleMatch {
					permil := int32(math.Round(r.Similarity * 1000))
					external.MetaMatch = &permil
				}

				err = db.UpdateGallery(gallery, tags, external)
				if err != nil {
					log.Debugf("Couldn't tag gallery: %s. Message: %s", gallery.ArchivePath, err)
					continue
				}

				if r.MetaTitleMatch {
					log.Info("Exact match based on meta titles: ", r.MatchedArchivePath)
				} else {
					log.Infof("Fuzzy match (%s): %s", fmt.Sprintf("%.2f", r.Similarity), r.MatchedArchivePath)
				}
			}
		}
	}
}
