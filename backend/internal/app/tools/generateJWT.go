package tools

import (
	"sync"
	"time"
	"errors"
	"context"
	"crypto/ecdsa"

	"github.com/OscarVillanueva/goapi/internal/platform"

	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
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
	ctx := context.Background()

	var privateKey *ecdsa.PrivateKey
	var publicKey *ecdsa.PublicKey

	privateStr, prErr := platform.GetSecret("private-key", ctx)
	publicStr, pubErr := platform.GetSecret("public-key", ctx)

	if prErr != nil || pubErr != nil {
		log.Warning("Read error: ", prErr)
		log.Warning("Read error: ", pubErr)

		if result, err := generateJWTKeys(); err == nil {
			privateKey = result.Sign
			publicKey = result.Verify

			handleSaveCredentials(result.Public, result.Private, ctx)
		}
	} else {
		privateKey, prErr = parsePrivatePemToKey(privateStr)
		publicKey, pubErr = parsePublicPemToKey(publicStr)

		if prErr != nil || pubErr != nil {
			log.Warning("Parse error: ", prErr)
			log.Warning("Parse error: ", pubErr)

			if result, err := generateJWTKeys(); err == nil {
				privateKey = result.Sign
				publicKey = result.Verify

				handleSaveCredentials(result.Public, result.Private, ctx)
			}
		}
	}

	manager.PublicKey = publicKey
	manager.PrivateKey = privateKey
}

func handleSaveCredentials(public string, private string, ctx context.Context)  {
	errPr := platform.SaveSecret("private-key", private, ctx)
	errPub := platform.SaveSecret("public-key", public, ctx)

	if errPr != nil || errPub != nil {
		log.Warning("Couldn't update the Credentials")
	} else {
		log.Info("Saved New Credentials")
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
			"exp": expirationDate.Unix(),
		})

	return token.SignedString(manager.PrivateKey)
}
