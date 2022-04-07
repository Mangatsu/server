package model

import (
	"time"
)

type User struct {
	UUID      string    `db:"uuid"`
	Username  string    `db:"username"`
	Password  string    `db:"password"`
	Role      int32     `db:"role"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
