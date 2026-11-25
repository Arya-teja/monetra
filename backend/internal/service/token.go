package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenManager issues and verifies HS256 access tokens whose subject is the user ID.
type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: ttl}
}

type AccessToken struct {
	Value     string
	ExpiresAt time.Time
}

// claims adds the user's token version, so a logout can revoke earlier tokens.
type claims struct {
	jwt.RegisteredClaims
	Version int `json:"ver"`
}

func (m *TokenManager) Issue(userID string, version int) (AccessToken, error) {
	now := time.Now()
	expiresAt := now.Add(m.ttl)
	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		Version: version,
	}

	value, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(m.secret)
	if err != nil {
		return AccessToken{}, fmt.Errorf("sign token: %w", err)
	}
	return AccessToken{Value: value, ExpiresAt: expiresAt}, nil
}

// Verify returns the user ID and token version of a valid, unexpired token.
func (m *TokenManager) Verify(token string) (string, int, error) {
	c := &claims{}
	_, err := jwt.ParseWithClaims(token, c, func(*jwt.Token) (any, error) {
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil {
		return "", 0, fmt.Errorf("parse token: %w", err)
	}
	return c.Subject, c.Version, nil
}
