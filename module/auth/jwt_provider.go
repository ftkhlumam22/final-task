package authmodule

import (
	"time"

	"final-task/helper"
	"final-task/model"
)

func NewJWTTokenProvider(jwtManager *model.JWTManager) TokenProvider {
	return &helperTokenProvider{
		jwtManager: jwtManager,
	}
}

func (tokenProvider *helperTokenProvider) GenerateAccessToken(userID int64) (string, error) {
	return helper.GenerateAccessToken(tokenProvider.jwtManager, userID)
}

func (tokenProvider *helperTokenProvider) GenerateRefreshToken(userID int64) (string, string, error) {
	return helper.GenerateRefreshToken(tokenProvider.jwtManager, userID)
}

func (tokenProvider *helperTokenProvider) VerifyToken(tokenString string, expectedTokenType string) (*model.CustomClaims, error) {
	return helper.VerifyToken(tokenProvider.jwtManager, tokenString, expectedTokenType)
}

func (tokenProvider *helperTokenProvider) RefreshTTL() time.Duration {
	return helper.RefreshTTL(tokenProvider.jwtManager)
}
