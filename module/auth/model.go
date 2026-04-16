package authmodule

import (
	"time"

	"final-task/model"
)

type AuthService struct {
	userSQLRepository   UserSQLRepository
	tokenProvider       TokenProvider
	userCacheRepository UserCacheRepository
}

type UserSQLRepository interface {
	CreateUser(userName string, userEmail string, passwordHash string) (model.User, error)
	FindUserByEmail(userEmail string) (model.User, error)
}

type UserCacheRepository interface {
	Set(cacheKey string, cacheValue string, cacheTTL time.Duration) error
	Exists(cacheKey string) (bool, error)
	Delete(cacheKey string) error
}

type TokenProvider interface {
	GenerateAccessToken(userID int64) (string, error)
	GenerateRefreshToken(userID int64) (token string, tokenID string, err error)
	VerifyToken(tokenString string, expectedTokenType string) (model.CustomClaims, error)
	RefreshTTL() time.Duration
}

type helperTokenProvider struct {
	jwtManager model.JWTManager
}
