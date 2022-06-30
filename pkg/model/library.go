package model

type Library struct {
	ID     int32  `db:"id"`
	Path   string `db:"path"`
	Layout string `db:"layout"`
}

type CombinedLibrary struct {
	Library
	Galleries []Gallery `db:"gallery"`
}
