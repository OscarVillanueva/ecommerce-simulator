package requests

type Login struct {
	Email string `json:"email"`
	Token string `json:"token"`
}
