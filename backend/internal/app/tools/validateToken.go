package tools

import (
	"fmt"
	"errors"
	"context"

	"github.com/golang-jwt/jwt/v5"
)

func IsValidToken(tokenString string, ctx context.Context) (map[string]any, error)  {
	manager := getKeyManager(ctx)

	if manager.PublicKey == nil {
		return nil, errors.New("Public key not found")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check if the token has a method pointer of type ECDSA
		// variable.(Type) is called Type assertion like a typeof of JS
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return manager.PublicKey, nil
	})

	if err != nil {
		return nil, err
	}

	// Verify that the claims is a map
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("Invalid token")
}
