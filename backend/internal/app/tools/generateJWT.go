package tools

import (
	"time"

	"github.com/joho/godotenv"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(expirationDate time.Time, data map[string]any) (string, error) {
	if err := godotenv.Load(); err != nil {
		return "", err
	}

	result, err := generateJWTKeys()

	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodES256,
		jwt.MapClaims {
			"data": data,
			"expiration_date": expirationDate.String(),
		})

	return token.SignedString(result.Sign)
}
