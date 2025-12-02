package requests

type VerifyAccount struct {
	Email string `json:"email"`
	Token string `json:"token"`
}
