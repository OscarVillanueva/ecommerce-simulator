package dao

import "time"

type Product struct {
	Uuid string
	Name string
	Price float32
	Quantity int32
	Image *string
	BelongsTo string
	CreatedAt time.Time
	UpdatedAt *time.Time
}
