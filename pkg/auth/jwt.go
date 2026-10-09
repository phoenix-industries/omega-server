package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultJWTIssuer       = "omega"
	DefaultJWTDuration     = 10 * time.Minute
	DefaultSessionDuration = 365 * 24 * time.Hour
)

type JWTClaims struct {
	jwt.RegisteredClaims
	Role Role `json:"role"`
}

func (c *JWTClaims) Valid() error {
	if c.Role == "" {
		return errors.New("invalid token: role is empty")
	}
	return nil
}

// GenerateJWT generates a JWT with the HS256 method.
func (a *Auth) GenerateJWT(subject, audience string, role Role) (string, error) {
	claims := JWTClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    defaultJWTIssuer,
			Subject:   subject,
			Audience:  []string{audience},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(DefaultJWTDuration)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := t.SignedString(a.jwtSecret)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return ss, nil
}

func (a *Auth) ParseJWT(token string) (*jwt.Token, error) {
	t, err := jwt.ParseWithClaims(token, &JWTClaims{}, a.jwtParseKeyFunc, jwt.WithLeeway(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	} else if !t.Valid {
		return nil, errors.New("invalid token")
	} else if _, ok := t.Claims.(*JWTClaims); !ok {
		return nil, errors.New("invalid token claims")
	}

	issuer, err := t.Claims.GetIssuer()
	if err != nil {
		return nil, fmt.Errorf("invalid token issuer: %w", err)
	} else if issuer != defaultJWTIssuer {
		return nil, errors.New("invalid token issuer")
	}

	return t, nil
}

func (a *Auth) GetJWTClaims(token *jwt.Token) (*JWTClaims, error) {
	tokenClaims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	return tokenClaims, nil
}

func (a *Auth) jwtParseKeyFunc(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, errors.New("invalid token")
	}
	return a.jwtSecret, nil
}
