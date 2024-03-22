package db

import (
	"github.com/Mangatsu/server/pkg/log"
	"go.uber.org/zap"
	"reflect"
	"strings"
	"time"

	"github.com/Mangatsu/server/pkg/types/sqlite/model"
	. "github.com/Mangatsu/server/pkg/types/sqlite/table"
	. "github.com/go-jet/jet/v2/sqlite"
)

// SanitizeString returns the pointer of the string with all leading and trailing white space removed.
// If the string is empty, it returns nil.
func SanitizeString(content *string) *string {
	if content == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*content)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// ValidateGallery returns updated and sanitized gallery model. For external API use.
func ValidateGallery(gallery model.Gallery, newGallery model.Gallery, now time.Time) (model.Gallery, ColumnList) {
	// Title cannot be empty or nil
	title := strings.TrimSpace(newGallery.Title)
	if title == "" {
		title = gallery.Title
	}
	gallery.Title = title

	// Boolean fields
	gallery.Nsfw = newGallery.Nsfw
	gallery.Hidden = newGallery.Hidden
	gallery.Translated = newGallery.Translated

	// Nullable string fields
	gallery.TitleNative = SanitizeString(newGallery.TitleNative)
	gallery.TitleTranslated = SanitizeString(newGallery.TitleTranslated)
	gallery.Category = SanitizeString(newGallery.Category)
	gallery.Released = SanitizeString(newGallery.Released)
	gallery.Series = SanitizeString(newGallery.Series)
	gallery.Language = SanitizeString(newGallery.Language)

	gallery.UpdatedAt = now

	return gallery, ColumnList{
		Gallery.Title,
		Gallery.TitleNative,
		Gallery.TitleTranslated,
		Gallery.Category,
		Gallery.Released,
		Gallery.Series,
		Gallery.Language,
		Gallery.Nsfw,
		Gallery.Hidden,
		Gallery.Translated,
		Gallery.UpdatedAt,
	}
}

// ValidateReference returns updated and sanitized gallery reference. For external API use.
func ValidateReference(reference model.Reference) model.Reference {
	reference.Urls = SanitizeString(reference.Urls)
	reference.ExhToken = SanitizeString(reference.ExhToken)

	return reference
}

// ValidateGalleryInternal returns updated and sanitized gallery model and column list.
// Non-empty and non-negative values are preferred.
// For internal scanner use.
func ValidateGalleryInternal(newGallery model.Gallery, now time.Time) (model.Gallery, ColumnList) {
	galleryModel := model.Gallery{}
	galleryUpdateColumnList := ColumnList{}

	// Title cannot be empty or nil
	if value := strings.TrimSpace(newGallery.Title); value != "" {
		galleryUpdateColumnList = append(galleryUpdateColumnList, Gallery.Title)
		galleryModel.Title = value
	}

	galleryUpdateColumnList = append(galleryUpdateColumnList, Gallery.Nsfw)
	galleryModel.Nsfw = newGallery.Nsfw

	galleryUpdateColumnList = append(galleryUpdateColumnList, Gallery.Hidden)
	galleryModel.Hidden = newGallery.Hidden

	if value := SanitizeString(newGallery.TitleNative); value != nil {
		galleryUpdateColumnList = append(galleryUpdateColumnList, Gallery.TitleNative)
		galleryModel.TitleNative = value
	}
	if value := SanitizeString(newGallery.TitleTranslated); value != nil {
		galleryUpdateColumnList = append(galleryUpdateColumnList, Gallery.TitleTranslated)
		galleryModel.TitleTranslated = value
	}
	if value := SanitizeString(newGallery.Category); value != nil {
		galleryUpdateColumnList = append(galleryUpdateColumnList, Gallery.Category)
		galleryModel.Category = value
	}
	if value := SanitizeString(newGallery.Released); value != nil {
		galleryUpdateColumnList = append(galleryUpdateColumnList, Gallery.Released)
		galleryModel.Released = value
	}
	if value := SanitizeString(newGallery.Series); value != nil {
		galleryUpdateColumnList = append(galleryUpdateColumnList, Gallery.Series)
		galleryModel.Series = value
	}
	if value := SanitizeString(newGallery.Language); value != nil {
		galleryUpdateColumnList = append(galleryUpdateColumnList, Gallery.Language)
		galleryModel.Language = value
	}
	if newGallery.Translated != nil {
		galleryUpdateColumnList = append(galleryUpdateColumnList, Gallery.Translated)
		galleryModel.Translated = newGallery.Translated
	}

	galleryUpdateColumnList = append(galleryUpdateColumnList, Gallery.Nsfw)
	galleryUpdateColumnList = append(galleryUpdateColumnList, Gallery.Hidden)

	if newGallery.ImageCount != nil && *newGallery.ImageCount >= 0 {
		galleryUpdateColumnList = append(galleryUpdateColumnList, Gallery.ImageCount)
		galleryModel.ImageCount = newGallery.ImageCount
	}
	if newGallery.ArchiveSize != nil && *newGallery.ArchiveSize >= 0 {
		galleryUpdateColumnList = append(galleryUpdateColumnList, Gallery.ArchiveSize)
		galleryModel.ArchiveSize = newGallery.ArchiveSize
	}
	if value := SanitizeString(newGallery.ArchiveHash); value != nil {
		galleryUpdateColumnList = append(galleryUpdateColumnList, Gallery.ArchiveHash)
		galleryModel.ArchiveHash = value
	}

	galleryUpdateColumnList = append(galleryUpdateColumnList, Gallery.UpdatedAt)
	galleryModel.UpdatedAt = now

	return galleryModel, galleryUpdateColumnList
}

// ValidateReferenceInternal returns updated and sanitized gallery reference.
// Non-empty values are preferred.
// For internal scanner use.
func ValidateReferenceInternal(reference model.Reference) model.Reference {
	referenceModel := model.Reference{}

	referenceModel.MetaInternal = reference.MetaInternal

	if reference.MetaTitleHash != nil && *reference.MetaTitleHash != "" {
		referenceModel.MetaTitleHash = reference.MetaTitleHash
	}

	if value := SanitizeString(reference.MetaPath); value != nil {
		referenceModel.MetaPath = value
	}

	if reference.MetaMatch != nil {
		referenceModel.MetaPath = reference.MetaPath
	}

	if value := SanitizeString(reference.Urls); value != nil {
		referenceModel.Urls = value
	}

	if reference.ExhGid != nil {
		referenceModel.ExhGid = reference.ExhGid
	}

	if value := SanitizeString(reference.ExhToken); value != nil {
		referenceModel.ExhToken = value
	}

	if reference.AnilistID != nil && *reference.AnilistID > 0 {
		referenceModel.AnilistID = reference.AnilistID
	}

	return referenceModel
}

// ConstructExpressions constructs expressions from a reference model.
func ConstructExpressions(model interface{}) ([]Expression, ColumnList, error) {
	v := reflect.ValueOf(model)
	typeOfS := v.Type()
	values := make([]Expression, 0)

	referenceColumnList := ColumnList{}

	for i := 0; i < v.NumField(); i++ {
		var value Expression

		field := v.Field(i)
		kind := field.Kind()
		if kind == reflect.Ptr || kind == reflect.Interface {
			if field.IsNil() {
				continue
			}
			field = field.Elem()
			kind = field.Kind()
		}

		switch kind {
		case reflect.String:
			if v.Field(i).IsZero() {
				continue
			}
			value = String(field.String())
		case reflect.Bool:
			value = Bool(field.Bool())
		case reflect.Int:
			value = Int(field.Int())
		case reflect.Int32:
			value = Int32(int32(field.Int()))
		case reflect.Int64:
			value = Int64(field.Int())
		default:
			log.Z.Error("unsupported reflect type",
				zap.String("type", v.Field(i).Kind().String()),
				zap.String("field", v.Type().Field(i).Name),
			)
			continue
		}

		columName := ConstructColumn(typeOfS.Field(i).Name)
		if columName != nil {
			//log.Z.Debug("constructed column",
			//	zap.String("column", columName.Name()),
			//	zap.Any("value", value),
			//)

			referenceColumnList = append(referenceColumnList, columName)
			values = append(values, value)
		}
	}

	return values, referenceColumnList, nil
}

func ConstructColumn(columnName string) Column {
	switch columnName {
	case "GalleryUUID":
		return Reference.GalleryUUID
	case "MetaInternal":
		return Reference.MetaInternal
	case "MetaPath":
		return Reference.MetaPath
	case "MetaMatch":
		return Reference.MetaMatch
	case "Urls":
		return Reference.Urls
	case "ExhGid":
		return Reference.ExhGid
	case "ExhToken":
		return Reference.ExhToken
	case "AnilistID":
		return Reference.AnilistID
	case "MetaTitleHash":
		return Reference.MetaTitleHash
	default:
		log.Z.Error("unsupported column name", zap.String("column", columnName))
		return nil
	}
}
