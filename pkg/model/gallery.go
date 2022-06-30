package model

import (
	"time"
)

type Gallery struct {
	UUID            string    `db:"uuid"`
	LibraryID       int32     `db:"library_id"`
	ArchivePath     string    `db:"archive_path"`
	Title           string    `db:"title"`
	TitleNative     *string   `db:"title_native"`
	TitleTranslated *string   `db:"title_translated"`
	Category        *string   `db:"category"`
	Series          *string   `db:"series"`
	Released        *string   `db:"released"`
	Language        *string   `db:"language"`
	Translated      *bool     `db:"translated"`
	Nsfw            bool      `db:"nsfw"`
	Hidden          bool      `db:"hidden"`
	ImageCount      *int32    `db:"image_count"`
	ArchiveSize     *int32    `db:"archive_size"`
	ArchiveHash     *string   `db:"archive_hash"`
	Thumbnail       *string   `db:"thumbnail"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}
