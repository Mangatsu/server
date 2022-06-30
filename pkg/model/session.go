package model

import (
	"time"
)

type Session struct {
	ID        string     `db:"id"`
	UserUUID  string     `db:"user_uuid"`
	Name      *string    `db:"name"`
	ExpiresAt *time.Time `db:"expires_at"`
}
