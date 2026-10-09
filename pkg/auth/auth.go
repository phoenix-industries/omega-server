// Package auth provides authentication functionality.
package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
)

type Auth struct {
	jwtSecret      []byte
	passwordSecret []byte
}

func New(jwtSecret []byte, passwordSecret []byte) (*Auth, error) {
	if len(jwtSecret) != 32 {
		return nil, errors.New("jwt secret must be exactly 32 bytes")
	}
	if len(passwordSecret) != 32 {
		return nil, errors.New("password signing key must be exactly 32 bytes")
	}
	return &Auth{
		jwtSecret:      jwtSecret,
		passwordSecret: passwordSecret,
	}, nil
}

func NewFromEnv() (*Auth, error) {
	jwtSecret, err := base64.StdEncoding.DecodeString(os.Getenv("AUTH_JWT_SECRET"))
	if err != nil {
		return nil, errors.New("failed to decode jwt secret")
	}
	passwordSecret, err := base64.StdEncoding.DecodeString(os.Getenv("AUTH_PASSWORD_SECRET"))
	if err != nil {
		return nil, fmt.Errorf("failed to decode signer key: %v", err)
	}
	return New(jwtSecret, passwordSecret)
}
