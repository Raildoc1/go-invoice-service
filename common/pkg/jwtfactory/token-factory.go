package jwtfactory

import (
	"fmt"
	"time"

	"github.com/go-chi/jwtauth/v5"
)

type TokenFactory struct {
	tokenAuth           *jwtauth.JWTAuth
	tokenExpirationTime time.Duration
}

type Config struct {
	Algorithm      string
	Secret         string
	ExpirationTime time.Duration
}

func New(cfg Config) *TokenFactory {
	tokenAuth := jwtauth.New(cfg.Algorithm, []byte(cfg.Secret), nil)
	return &TokenFactory{
		tokenAuth:           tokenAuth,
		tokenExpirationTime: cfg.ExpirationTime,
	}
}

func (tf *TokenFactory) GetJWTAuth() *jwtauth.JWTAuth {
	return tf.tokenAuth
}

func (tf *TokenFactory) Generate(extraClaims map[string]string) (string, error) {
	timeNow := time.Now()
	claims := map[string]any{
		"exp": timeNow.Add(tf.tokenExpirationTime).Unix(),
		"iat": timeNow.Unix(),
	}
	for k, v := range extraClaims {
		claims[k] = v
	}
	_, tokenString, err := tf.tokenAuth.Encode(claims)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return tokenString, nil
}
