package metadata

import (
	"encoding/json"
	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/constants"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/log"
	"github.com/Mangatsu/server/pkg/types/sqlite/model"
	"github.com/Mangatsu/server/pkg/utils"
	"go.uber.org/zap"
	"os"
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
		log.Z.Error("failed to unmarshal x metadata", zap.String("err", err.Error()))
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

type FuzzyResult struct {
	MetaTitleMatch     bool
	Similarity         float64
	MatchedArchivePath string
	RelativeMetaPath   string
}

// fuzzyMatchExternalMeta tries to find the metadata file besides it (fuzzy match).
func fuzzyMatchExternalMeta(archivePath string, libraryPath string, f os.DirEntry) (FuzzyResult, XMetadata) {
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

	metaData, err := utils.ReadJSON(config.BuildPath(onlyDir, f.Name()))
	if err != nil {
		log.Z.Debug("failed to unmarshal x metadata", zap.String("err", err.Error()))
		return fuzzyResult, XMetadata{}
	}

	exhGallery, err := unmarshalExhJSON(metaData)
	if err != nil {
		log.Z.Debug("could not unmarshal exhGallery", zap.String("err", err.Error()))
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
	archiveNoExt := constants.ArchiveExtensions.ReplaceAllString(path.Base(relativeArchivePath), "")

	archiveSimilarity := utils.Similarity(archiveNoExt, metaNoExt)
	titleSimilarity := utils.Similarity(archiveNoExt, *exhGallery.GalleryInfo.Title)
	titleNativeSimilarity := utils.Similarity(archiveNoExt, *exhGallery.GalleryInfo.TitleOriginal)
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

// ParseX parses x JSON file (x: https://github.com/dnsev-h/x).
func ParseX(metaData []byte, metaPath string, archivePath string, internal bool) (model.Gallery, []model.Tag, model.Reference, error) {
	exhGallery, err := unmarshalExhJSON(metaData)
	if err != nil {
		return model.Gallery{}, nil, model.Reference{}, err
	}

	gallery, tags, reference := convertExh(exhGallery, archivePath, metaPath, internal)

	return gallery, tags, reference, nil
}
