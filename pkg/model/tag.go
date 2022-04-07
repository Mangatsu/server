package model

type Tag struct {
	ID        int32  `db:"id"`
	Namespace string `db:"namespace"`
	Name      string `db:"name"`
}
