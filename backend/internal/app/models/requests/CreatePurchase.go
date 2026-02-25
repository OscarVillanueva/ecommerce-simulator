package requests

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

type CreatePurchase struct {
	Product string `json:"product"`
	Quantity int32 `json:"quantity"`
}

func (c *CreatePurchase) Validate() error {
	trimedString := strings.TrimSpace(c.Product)

	if trimedString == "" || len(trimedString) < 35  {
		return errors.New("Product uuid is required")
	}

	if _, err := uuid.Parse(trimedString); err != nil {
		return errors.New("Product uuid is invalid")
	}

	if c.Quantity < 0 {
		return errors.New("Quantity cannot be negative")
	}

	return nil
}
