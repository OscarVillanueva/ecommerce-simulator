package tools

import (
	"os"
	"time"
	"crypto/ecdsa"

	"github.com/joho/godotenv"
	"github.com/golang-jwt/jwt/v5"
)

const JWT_KEY = "PRIVATE_JWT_KEY"
const VERIFY_KEY = "PUBLIC_JWT_KEY"

func GenerateJWT(expirationDate time.Time, data map[string]any) (string, error) {
	if err := godotenv.Load(); err != nil {
		return "", err
	}

	var private *ecdsa.PrivateKey
	key := os.Getenv(JWT_KEY)

	if key != "" {
		pem, err := parsePrivatePemToKey(key)
		if err != nil { return "", err }

		private = pem
	} else {
		
		result, err := generateJWTKeys()
		if err != nil {
			return "", err
		}

		if err := updateEnvFile(JWT_KEY, result.Private); err != nil {
			return "", err
		}

		_ = updateEnvFile(VERIFY_KEY, result.Public)

		private = result.Sign
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodES256,
		jwt.MapClaims {
			"data": data,
			"expiration_date": expirationDate.String(),
		})

	return token.SignedString(private)
}
