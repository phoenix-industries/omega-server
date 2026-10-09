package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/scrypt"
)

const (
	scryptMemCost = 1 << 16
	scryptRounds  = 8
	scryptP       = 1
	scryptKeyLen  = 32
	scryptSaltLen = 16
	scryptHashID  = "scrypt"
	scryptHashSep = "$"
)

func (a *Auth) HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password is empty")
	}

	salt, err := a.generateSalt()
	if err != nil {
		return "", err
	}

	key, err := a.scryptHash([]byte(password), salt, scryptMemCost, scryptRounds, scryptP, scryptKeyLen)
	if err != nil {
		return "", err
	}

	keyB64 := base64.RawStdEncoding.EncodeToString(key)
	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	params := fmt.Sprintf("N=%d,r=%d,p=%d", scryptMemCost, scryptRounds, scryptP)
	hash := scryptHashSep + scryptHashID + scryptHashSep + params + scryptHashSep + saltB64 + scryptHashSep + keyB64

	return hash, nil
}

func (a *Auth) VerifyPassword(password, hash string) (bool, error) {
	if password == "" {
		return false, errors.New("password is empty")
	}
	if hash == "" {
		return false, errors.New("hash is empty")
	}

	parts := strings.SplitN(hash, scryptHashSep, 5)
	if len(parts) != 5 || parts[0] != "" || parts[1] != scryptHashID {
		return false, errors.New("invalid hash format")
	}
	params, saltB64, keyB64 := parts[2], parts[3], parts[4]

	var n, r, p int
	if _, err := fmt.Sscanf(params, "N=%d,r=%d,p=%d", &n, &r, &p); err != nil {
		return false, fmt.Errorf("invalid %s params: %w", scryptHashID, err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false, fmt.Errorf("invalid salt encoding: %w", err)
	}

	hashedKey, err := base64.RawStdEncoding.DecodeString(keyB64)
	if err != nil {
		return false, fmt.Errorf("invalid key encoding: %w", err)
	}

	key, err := a.scryptHash([]byte(password), salt, n, r, p, len(hashedKey))
	if err != nil {
		return false, err
	}

	if len(key) != len(hashedKey) {
		return false, nil
	}

	return subtle.ConstantTimeCompare(key, hashedKey) == 1, nil
}

func (a *Auth) generateSalt() ([]byte, error) {
	salt := make([]byte, scryptSaltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

func (a *Auth) scryptHash(password, salt []byte, n, r, p, keyLen int) ([]byte, error) {
	hash, err := scrypt.Key(
		password,
		salt,
		n,
		r,
		p,
		keyLen,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	return hash, nil
}
