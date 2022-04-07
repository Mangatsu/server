package model

type GalleryTag struct {
	GalleryUUID string `db:"gallery_uuid"`
	TagID       int32  `db:"tag_id"`
}
