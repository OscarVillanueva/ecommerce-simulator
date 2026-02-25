package requests

import (
	"strings"
	"errors"
)

type CreateProduct struct {
	Name string `json:"name"`
	Price float32 `json:"price"`
	Quantity int32 `json:"quantity`
}

type RequestSchema interface {
	Validate() error
}

func (c *CreateProduct) Validate() error {
	trimedName := strings.TrimSpace(c.Name)

	if trimedName == "" || (len(trimedName) < 3 && len(trimedName) > 50) {
		return errors.New("The product name should be greater than 3 and less that 50")
	}
	if c.Price < 0 {
		return errors.New("Price cannot be negative")
	}
	if c.Quantity < 0 {
		return errors.New("Quantity cannot be negative")
	}

	return nil
}

