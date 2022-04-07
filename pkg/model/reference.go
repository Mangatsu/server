package model

type Reference struct {
	GalleryUUID  string  `db:"gallery_uuid"`
	MetaInternal bool    `db:"meta_internal"`
	MetaPath     *string `db:"meta_path"`
	MetaMatch    *int32  `db:"meta_match"`
	Urls         *string `db:"urls"`
	ExhGid       *int32  `db:"exh_gid"`
	ExhToken     *string `db:"exh_token"`
	AnilistID    *int32  `db:"anilist_id"`
}
