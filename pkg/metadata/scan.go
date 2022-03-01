package metadata

import (
	"errors"
	"fmt"
	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/library"
	"github.com/Mangatsu/server/pkg/types/model"
	"github.com/mholt/archiver/v4"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"io/ioutil"
	"math"
	"path/filepath"
)

type MetaType string

const (
	XMeta    MetaType = "xmeta"
	HathMeta          = "hathmeta"
	EHDLMeta          = "ehdlmeta"
)

// matchInternalMeta reads the internal metadata (info.json, info.txt or galleryinfo.txt) from the given archive.
func matchInternalMeta(metaTypes map[MetaType]bool, fullArchivePath string) ([]byte, string, MetaType) {
	fsys, err := archiver.FileSystem(fullArchivePath)
	if err != nil {
		log.Error("Error opening archive: ", err)
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
		log.Debug("Couldn't read external metadata: ", err)
		return nil, ""
	}

	return metaData, config.RelativePath(libraryPath, externalJSON)
}

// ParseMetadata scans all libraries for metadata files (json, txt).
func ParseMetadata(metaTypes map[MetaType]bool) {
	libraries, err := db.GetLibraries()
	if err != nil {
		log.Error("Libraries could not be retrieved to parse meta files: ", err)
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
						log.Debug("Couldn't parse X meta: ", err)
						continue
					}
				case EHDLMeta:
					if newGallery, tags, err = ParseEHDL(metaPath); err != nil {
						log.Debug("Couldn't parse EHDL meta: ", err)
						continue
					}
				case HathMeta:
					if newGallery, tags, err = ParseHath(metaPath); err != nil {
						log.Debug("Couldn't parse Hath meta: ", err)
						continue
					}
				}

				if err = db.UpdateGallery(newGallery, tags, &reference, true); err != nil {
					log.Debugf("Couldn't tag gallery: %s. Message: %s", gallery.ArchivePath, err)
					continue
				}
			}
		}
	}

	// Fuzzy parsing for all archives that didn't have an exact match.
	for _, noMatch := range archivesNoMatch {
		onlyDir := filepath.Dir(noMatch.fullPath)
		files, err := ioutil.ReadDir(onlyDir)
		if err != nil {
			log.Debug("Couldn't read dir while fuzzy matching: ", err)
		}

		for _, f := range files {
			r, exhGallery := fuzzyMatchExternalMeta(noMatch.fullPath, noMatch.libraryPath, f)

			if r.MatchedArchivePath != "" && r.MetaTitleMatch || r.Similarity > 0.70 {
				gallery, tags, reference := convertExh(exhGallery, r.MatchedArchivePath, r.RelativeMetaPath, false)

				if !r.MetaTitleMatch {
					permil := int32(math.Round(r.Similarity * 1000))
					reference.MetaMatch = &permil
				}

				err = db.UpdateGallery(gallery, tags, &reference, true)
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
