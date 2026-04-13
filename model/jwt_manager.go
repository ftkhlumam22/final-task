package model

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	Config JWTConfig
}

func NewJWTManager(jwtConfig JWTConfig) *JWTManager {
	return &JWTManager{Config: jwtConfig}
}

func (jwtManager *JWTManager) RefreshTTL() time.Duration {
	return jwtManager.Config.RefreshTokenTTL
}

func (jwtManager *JWTManager) GenerateAccessToken(userID int64) (string, error) {
	return jwtManager.generateToken(userID, TokenTypeAccess, "", jwtManager.Config.AccessTokenTTL)
}

func (jwtManager *JWTManager) GenerateRefreshToken(userID int64) (token string, tokenID string, tokenError error) {
	tokenID, tokenError = newTokenID()
	if tokenError != nil {
		return "", "", tokenError
	}

	token, tokenError = jwtManager.generateToken(userID, TokenTypeRefresh, tokenID, jwtManager.Config.RefreshTokenTTL)
	if tokenError != nil {
		return "", "", tokenError
	}

	return token, tokenID, nil
}

func (jwtManager *JWTManager) VerifyToken(tokenString string, expectedTokenType string) (*CustomClaims, error) {
	parsedToken, parseError := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New(MessageUnexpectedSigning)
		}
		return []byte(jwtManager.Config.SecretKey), nil
	})
	if parseError != nil {
		return nil, parseError
	}

	claims, isValidClaimType := parsedToken.Claims.(*CustomClaims)
	if !isValidClaimType || !parsedToken.Valid {
		return nil, errors.New(MessageInvalidToken)
	}

	if claims.TokenType != expectedTokenType {
		return nil, errors.New(MessageInvalidTokenType)
	}

	return claims, nil
}

func (jwtManager *JWTManager) generateToken(userID int64, tokenType string, tokenID string, tokenTTL time.Duration) (string, error) {
	currentTime := time.Now()

	customClaims := CustomClaims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Issuer:    jwtManager.Config.Issuer,
			Subject:   strconv.FormatInt(userID, 10),
			ExpiresAt: jwt.NewNumericDate(currentTime.Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(currentTime),
			NotBefore: jwt.NewNumericDate(currentTime),
		},
	}

	generatedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims)
	return generatedToken.SignedString([]byte(jwtManager.Config.SecretKey))
}

func newTokenID() (string, error) {
	randomBytes := make([]byte, 16)
	if _, readError := rand.Read(randomBytes); readError != nil {
		return "", readError
	}

	return hex.EncodeToString(randomBytes), nil
}
