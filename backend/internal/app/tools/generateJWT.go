package tools

import (
	"errors"
	"sync"
	"time"
	"crypto/ecdsa"

	"github.com/golang-jwt/jwt/v5"
)


type JWTKeyManager struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey *ecdsa.PublicKey
}

var (
	instance *JWTKeyManager
	once sync.Once
)

func getKeyManager() *JWTKeyManager  {
	once.Do(func() {
		instance = &JWTKeyManager{}
		instance.initialize()
	})

	return instance
}

func (manager *JWTKeyManager) initialize()  {
	if result, err := generateJWTKeys(); err == nil {
		manager.PublicKey = result.Verify
		manager.PrivateKey = result.Sign
	}
}

func GenerateJWT(expirationDate time.Time, data map[string]any) (string, error) {
	manager := getKeyManager()

	if manager.PrivateKey == nil {
		return "", errors.New("We couldn't sing the token")
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodES256,
		jwt.MapClaims {
			"data": data,
			"expiration_date": expirationDate.String(),
		})

	return token.SignedString(manager.PrivateKey)
}
