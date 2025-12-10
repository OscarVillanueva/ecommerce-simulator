package tools

import (
	"errors"
	"crypto/x509"
	"crypto/ecdsa"
	"encoding/pem"
)

func parsePrivatePemToKey(private string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(private))
	if block == nil {
		return nil, errors.New("Failed to parse PEM block containing the key")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key, err
}
