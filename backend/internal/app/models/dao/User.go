package dao

import "time"

type User struct {
	Uuid string
	Name string
	Email string
	Verified bool
	CreatedAt time.Time
}

