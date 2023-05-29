package metadata

import (
	"errors"
	"io/fs"
	"io/ioutil"
	"math"
	"path/filepath"

	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/library"
	"github.com/Mangatsu/server/pkg/log"
	"github.com/Mangatsu/server/pkg/types/model"
	"github.com/mholt/archiver/v4"
	"go.uber.org/zap"
)

type MetaType string

const (
	XMeta    MetaType = "xmeta"
	HathMeta          = "hathmeta"
	EHDLMeta          = "ehdlmeta"
)

// matchInternalMeta reads the internal metadata (info.json, info.txt or galleryinfo.txt) from the given archive.
func matchInternalMeta(metaTypes map[MetaType]bool, fullArchivePath string) ([]byte, string, MetaType) {
	fsys, err := archiver.FileSystem(nil, fullArchivePath)
	if err != nil {
		log.Z.Error("could not open archive",
			zap.String("path", fullArchivePath),
			zap.String("err", err.Error()))
		return nil, "", ""
	}

	var content []byte
	var filename string
	var metaType MetaType = ""

	err = fs.WalkDir(fsys, ".", func(s string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.Name() == "info.json" && metaTypes[XMeta] {
			metaType = XMeta
		} else if d.Name() == "info.txt" && metaTypes[EHDLMeta] {
			metaType = EHDLMeta
		} else if d.Name() == "galleryinfo.txt" && metaTypes[HathMeta] {
			metaType = HathMeta
		}

		if metaType != "" {
			filename = s
			content, err = library.ReadAll(fsys, s)
			if err != nil {
				return err
			}
			return errors.New("terminate walk")
		}
		return nil
	})

	return content, filename, metaType
}

// matchExternalMeta tries to find the metadata file besides it (exact match).
func matchExternalMeta(metaTypes map[MetaType]bool, fullArchivePath string, libraryPath string) ([]byte, string) {
	if !metaTypes[XMeta] {
		return nil, ""
	}

	externalJSON := library.ArchiveExtensions.ReplaceAllString(fullArchivePath, ".json")

	if !library.PathExists(externalJSON) {
		archivesNoMatch = append(archivesNoMatch, NoMatchPaths{libraryPath: libraryPath, fullPath: fullArchivePath})
		return nil, ""
	}

	metaData, err := library.ReadJSON(externalJSON)
	if err != nil {
		log.Z.Debug("could not read external metadata",
			zap.String("path", externalJSON),
			zap.String("err", err.Error()))
		return nil, ""
	}

	return metaData, config.RelativePath(libraryPath, externalJSON)
}

// ParseMetadata scans all libraries for metadata files (json, txt).
func ParseMetadata(metaTypes map[MetaType]bool) {
	libraries, err := db.GetLibraries()
	if err != nil {
		log.Z.Error("libraries could not be retrieved to parse meta files: ", zap.String("err", err.Error()))
		return
	}

	for _, galleryLibrary := range libraries {
		for _, gallery := range galleryLibrary.Galleries {
			fullPath := config.BuildLibraryPath(galleryLibrary.Path, gallery.ArchivePath)

			var metaData []byte
			var metaPath string
			internal := false

			// X, Hath, EHDL
			metaData, metaPath, metaType := matchInternalMeta(metaTypes, fullPath)
			if metaData != nil {
				internal = true
			}

			// X
			if !internal {
				metaData, metaPath = matchExternalMeta(metaTypes, fullPath, galleryLibrary.Path)
				metaType = XMeta
			}

			if metaData != nil {
				var newGallery model.Gallery
				var tags []model.Tag
				var reference model.Reference

				switch metaType {
				case XMeta:
					if newGallery, tags, reference, err = ParseX(metaData, metaPath, gallery.ArchivePath, internal); err != nil {
						log.Z.Debug("could not parse X meta",
							zap.String("path", metaPath),
							zap.String("err", err.Error()))
						continue
					}
				case EHDLMeta:
					if newGallery, tags, err = ParseEHDL(metaPath); err != nil {
						log.Z.Debug("could not parse EHDL meta",
							zap.String("path", metaPath),
							zap.String("err", err.Error()))
						continue
					}
				case HathMeta:
					if newGallery, tags, err = ParseHath(metaPath); err != nil {
						log.Z.Debug("could not parse Hath meta",
							zap.String("path", metaPath),
							zap.String("err", err.Error()))
						continue
					}
				}

				if err = db.UpdateGallery(newGallery, tags, reference, true); err != nil {
					log.Z.Debug("could not tag gallery",
						zap.String("path", gallery.ArchivePath),
						zap.String("err", err.Error()))
					continue
				}
			}
		}
	}

	// Fuzzy parsing for all archives that didn't have an exact match.
	for _, noMatch := range archivesNoMatch {
		onlyDir := filepath.Dir(noMatch.fullPath)
		files, err := ioutil.ReadDir(onlyDir) // TODO: Replace with os.ReadDir as this is deprecated as of Go 1.16
		if err != nil {
			log.Z.Debug("could not gallery read dir while fuzzy matching",
				zap.String("path", onlyDir),
				zap.String("err", err.Error()))
		}

		for _, f := range files {
			r, exhGallery := fuzzyMatchExternalMeta(noMatch.fullPath, noMatch.libraryPath, f)

			if r.MatchedArchivePath != "" && r.MetaTitleMatch || r.Similarity > 0.70 {
				gallery, tags, reference := convertExh(exhGallery, r.MatchedArchivePath, r.RelativeMetaPath, false)

				if !r.MetaTitleMatch {
					permil := int32(math.Round(r.Similarity * 1000))
					reference.MetaMatch = &permil
				}

				err = db.UpdateGallery(gallery, tags, reference, true)
				if err != nil {
					log.Z.Debug("could not tag gallery",
						zap.String("path", gallery.ArchivePath),
						zap.String("err", err.Error()))
					continue
				}

				if r.MetaTitleMatch {
					log.Z.Info("exact match based on meta titles", zap.String("path", r.MatchedArchivePath))
				} else {
					log.Z.Info("fuzzy match",
						zap.Float64("similarity", r.Similarity),
						zap.String("path", r.MatchedArchivePath))
				}
			}
		}
	}
}
