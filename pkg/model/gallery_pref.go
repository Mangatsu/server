package model

import (
	"time"
)

type GalleryPref struct {
	UserUUID      string    `db:"user_uuid"`
	GalleryUUID   string    `db:"gallery_uuid"`
	Progress      int32     `db:"progress"`
	FavoriteGroup *string   `db:"favorite_group"`
	UpdatedAt     time.Time `db:"updated_at"`
}
