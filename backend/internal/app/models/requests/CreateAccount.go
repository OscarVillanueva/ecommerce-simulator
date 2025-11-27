package requests

type CreateAccount struct {
	Name string `json:"name"`
	Email string `json:"email"`
}

