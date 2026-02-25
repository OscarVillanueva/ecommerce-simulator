package requests

import (
	"errors"
	"strings"
	"net/mail"
)

type VerifyAccount struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

func (verify VerifyAccount) Validate() error {
	trimedToken := strings.TrimSpace(verify.Token)
	trimedEmail := strings.TrimSpace(verify.Email)

	if trimedEmail == "" {
		return errors.New("The email is required")
	}

	if _, err := mail.ParseAddress(trimedEmail); err != nil {
		return errors.New("Invalid email")
	}
	
	if trimedToken == "" || len(trimedToken) != 6{
		return errors.New("The token is required")
	}

	return nil
}

