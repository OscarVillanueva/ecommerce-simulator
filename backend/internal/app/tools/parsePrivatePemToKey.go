package tools

import (
	"fmt"
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

func parsePublicPemToKey(public string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(public))
	if block == nil {
		return nil, errors.New("Failed to parse PEM block containing the key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
		case *ecdsa.PublicKey:
			return pub, nil
		default:
			return nil, fmt.Errorf("unknown type of public key: %T", pub)
	}
}
