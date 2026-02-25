package requests

import (
	"errors"
	"net/mail"
)

type CreateAccount struct {
	Name string `json:"name"`
	Email string `json:"email"`
}

func (c *CreateAccount) Validate() error {
	if c.Name == "" || (len(c.Name) < 50 && len(c.Name) > 3)  {
		return errors.New("The account name should be greater than 3 and less that 50")
	}

	if c.Email == "" {
		return errors.New("Quantity cannot be negative")
	}

	if _, err := mail.ParseAddress(c.Email); err != nil {
		return errors.New("Invalid email")
	}

	return nil
}

