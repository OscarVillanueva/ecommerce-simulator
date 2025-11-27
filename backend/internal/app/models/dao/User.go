package dao

import "time"

type User struct {
	Uuid string
	Name string
	Email string
	CreatedAt time.Time
}

