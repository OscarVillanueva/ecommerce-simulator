package dao

import "time"

type Magic struct {
	Token string
	ExpirationDate time.Time
	BelongsTo string
}

