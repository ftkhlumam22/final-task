package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	SecretKey       string
	Issuer          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type CustomClaims struct {
	UserID    int64  `json:"user_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}
