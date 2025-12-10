package tools

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
)

type JWTKeys struct {
	Sign *ecdsa.PrivateKey
	Private string
	Public string
}

func generateJWTKeys() (*JWTKeys, error)  {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
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
		Private: privateKeyString,
		Public: publicKeyString,
	}

	return &result, nil
}
