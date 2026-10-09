package auth

import (
	"fmt"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	idLength = 21
)

func GenerateID() (string, error) {
	id, err := gonanoid.New(idLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate id: %w", err)
	}
	return id, nil
}

func (a *Auth) GenerateID() (string, error) {
	return GenerateID()
}
