package dao

import "time"

type Purchase struct {
	Uuid string
	Product string
	Quantity int32
	Price float32
	PurchasedBy string
	CreatedAt time.Time
}
