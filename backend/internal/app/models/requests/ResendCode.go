package requests

import (
	"errors"
	"strings"
	"net/mail"
)

type ResendCode struct {
	Email string `json:"email"`
}

func (resend ResendCode) Validate() error {
	trimedEmail := strings.TrimSpace(resend.Email)

	if trimedEmail == "" {
		return errors.New("The email is required")
	}

	if _, err := mail.ParseAddress(trimedEmail); err != nil {
		return errors.New("Invalid email")
	}

	return nil
}
