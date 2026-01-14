package dao

import "time"

type Product struct {
	Uuid string `json:"uuid"`
	Name string `json:"name"`
	Price float32 `json:"price"`
	Quantity int32 `json:"quantity"`
	Image *string `json:"image"`
	BelongsTo string `json:"belongs_to"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}
