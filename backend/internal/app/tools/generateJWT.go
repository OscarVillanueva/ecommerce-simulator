package tools

import (
	"fmt"
	"sync"
	"time"
	"errors"
	"context"
	"crypto/ecdsa"

	"github.com/OscarVillanueva/goapi/internal/platform"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel"
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

func getKeyManager(ctx context.Context) *JWTKeyManager  {
	once.Do(func() {
		instance = &JWTKeyManager{}
		instance.initialize(ctx)
	})

	return instance
}

const GenJWT = "generate-jwt"

func (manager *JWTKeyManager) initialize(c context.Context)  {
	tr := otel.Tracer(GenJWT)
	ctx, span := tr.Start(c, fmt.Sprintf("%s-initialize", GenJWT))
	defer span.End()

	err := loadExistingKeys(manager, ctx)
	if err == nil {
		span.SetStatus(codes.Ok, "Loading JWT Keys successfully")
		return
	}

	span.RecordError(err)
	span.SetStatus(codes.Error, "Unable to load secrets: " + err.Error())

	result, err := generateJWTKeys(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Unable to create secrets: " + err.Error())
		return
	}


	manager.PublicKey = result.Verify
	manager.PrivateKey = result.Sign

	err = handleSaveCredentials(result.Public, result.Private, ctx)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	span.SetStatus(codes.Ok, "JWT keys generated")
}

func loadExistingKeys(manager *JWTKeyManager, ctx context.Context) error {
	privateStr, err := platform.GetSecret("private-key", ctx)
	if err != nil {
		return err
	}

	publicStr, err := platform.GetSecret("public-key", ctx)
	if err != nil {
		return err
	}

	priv, err := parsePrivatePemToKey(privateStr)
	if err != nil {
		return err
	}

	pub, err := parsePublicPemToKey(publicStr)
	if err != nil {
		return err
	}

	manager.PublicKey = pub
	manager.PrivateKey = priv
	return nil
}

func handleSaveCredentials(public string, private string, ctx context.Context) error {
	errPr := platform.SaveSecret("private-key", private, ctx)
	errPub := platform.SaveSecret("public-key", public, ctx)

	return errors.Join(errPr, errPub)
}

func GenerateJWT(expirationDate time.Time, data map[string]any, ctx context.Context) (string, error) {
	manager := getKeyManager(ctx)

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
