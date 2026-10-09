package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const (
	tokenLength = 32
	TokenType   = "Bearer"
	TokenPrefix = TokenType + " "
)

func GenerateToken() (string, error) {
	b := make([]byte, tokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (a *Auth) GenerateToken() (string, error) {
	return GenerateToken()
}
