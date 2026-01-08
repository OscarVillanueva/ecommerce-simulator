package requests

import "errors"

type CreateProduct struct {
	Name string `json:"name"`
	Price float32 `json:"price"`
	Quantity int32 `json:"quantity`
}

func (c *CreateProduct) Validate() error {
	if c.Name == "" {
		return errors.New("Product name is required")
	}
	if c.Price < 0 {
		return errors.New("Price cannot be negative")
	}
	if c.Quantity < 0 {
		return errors.New("Quantity cannot be negative")
	}

	return nil
}

