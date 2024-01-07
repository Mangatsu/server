package db

import (
	"strings"

	"github.com/Mangatsu/server/pkg/types/model"
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

// ValidateGallery returns updated and sanitized gallery model.
func ValidateGallery(gallery model.Gallery, newGallery model.Gallery) model.Gallery {
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

	return gallery
}

// ValidateReference returns updated and sanitized gallery reference.
func ValidateReference(reference model.Reference, newReference model.Reference) model.Reference {
	reference.Urls = SanitizeString(newReference.Urls)
	reference.ExhToken = SanitizeString(newReference.ExhToken)
	reference.ExhGid = newReference.ExhGid
	reference.AnilistID = newReference.AnilistID

	return reference
}
