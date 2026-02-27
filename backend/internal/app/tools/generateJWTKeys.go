package tools

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel"
)

type JWTKeys struct {
	Sign *ecdsa.PrivateKey
	Verify *ecdsa.PublicKey
	Private string
	Public string
}

func generateJWTKeys(ctx context.Context) (*JWTKeys, error)  {
	tr := otel.Tracer("generate-keys")
	_, span := tr.Start(ctx, "-create-keys")
	defer span.End()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// Encode Private Key to PEM
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: x509Encoded})
	privateKeyString := string(pemEncoded)

	// Encode Public Key to PEM
	publicKey := &privateKey.PublicKey
	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})
	publicKeyString := string(pemEncodedPub)

	result := JWTKeys{
		Sign: privateKey,
		Verify: publicKey,
		Private: privateKeyString,
		Public: publicKeyString,
	}

	span.SetStatus(codes.Ok, "Generated keys successfully")
	return &result, nil
}
