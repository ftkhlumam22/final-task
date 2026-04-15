package authmodule

import (
	"database/sql"
	"errors"
	"fmt"

	"final-task/helper"
	"final-task/model"
)

func NewAuthService(
	userSQLRepository UserSQLRepository,
	tokenProvider TokenProvider,
	userCacheRepository UserCacheRepository,
) *AuthService {
	return &AuthService{
		userSQLRepository:   userSQLRepository,
		tokenProvider:       tokenProvider,
		userCacheRepository: userCacheRepository,
	}
}

func (authService *AuthService) RegisterUser(
	userName string,
	userEmail string,
	userPassword string,
) (model.User, error) {
	passwordHash, err := helper.HashPassword(userPassword)
	if err != nil {
		return model.User{}, err
	}

	return authService.userSQLRepository.CreateUser(userName, userEmail, passwordHash)
}

func (authService *AuthService) LoginUser(
	userEmail string,
	userPassword string,
) (model.AuthResult, error) {
	foundUser, err := authService.userSQLRepository.FindUserByEmail(userEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AuthResult{}, model.ErrInvalidCredential
		}
		return model.AuthResult{}, err
	}

	if err = helper.ComparePassword(foundUser.PasswordHash, userPassword); err != nil {
		return model.AuthResult{}, model.ErrInvalidCredential
	}

	accessToken, err := authService.tokenProvider.GenerateAccessToken(foundUser.ID)
	if err != nil {
		return model.AuthResult{}, err
	}

	refreshToken, refreshTokenID, err := authService.tokenProvider.GenerateRefreshToken(foundUser.ID)
	if err != nil {
		return model.AuthResult{}, err
	}

	refreshTokenKey := authService.formatRefreshTokenKey(foundUser.ID, refreshTokenID)
	err = authService.userCacheRepository.Set(refreshTokenKey, "1", authService.tokenProvider.RefreshTTL())
	if err != nil {
		return model.AuthResult{}, err
	}

	return model.AuthResult{
		User:         foundUser,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (authService *AuthService) RefreshUserToken(
	refreshToken string,
) (model.AuthResult, error) {
	tokenClaims, err := authService.tokenProvider.VerifyToken(refreshToken, model.TokenTypeRefresh)
	if err != nil {
		return model.AuthResult{}, model.ErrInvalidRefresh
	}

	refreshTokenKey := authService.formatRefreshTokenKey(tokenClaims.UserID, tokenClaims.ID)
	refreshTokenExists, err := authService.userCacheRepository.Exists(refreshTokenKey)
	if err != nil {
		return model.AuthResult{}, err
	}
	if !refreshTokenExists {
		return model.AuthResult{}, model.ErrInvalidRefresh
	}

	err = authService.userCacheRepository.Delete(refreshTokenKey)
	if err != nil {
		return model.AuthResult{}, err
	}

	accessToken, err := authService.tokenProvider.GenerateAccessToken(tokenClaims.UserID)
	if err != nil {
		return model.AuthResult{}, err
	}

	newRefreshToken, newRefreshTokenID, err := authService.tokenProvider.GenerateRefreshToken(tokenClaims.UserID)
	if err != nil {
		return model.AuthResult{}, err
	}

	newRefreshTokenKey := authService.formatRefreshTokenKey(tokenClaims.UserID, newRefreshTokenID)
	err = authService.userCacheRepository.Set(newRefreshTokenKey, "1", authService.tokenProvider.RefreshTTL())
	if err != nil {
		return model.AuthResult{}, err
	}

	return model.AuthResult{
		User: model.User{
			ID: tokenClaims.UserID,
		},
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (authService *AuthService) LogoutUser(
	refreshToken string,
) error {
	tokenClaims, err := authService.tokenProvider.VerifyToken(refreshToken, model.TokenTypeRefresh)
	if err != nil {
		return model.ErrInvalidRefresh
	}

	refreshTokenKey := authService.formatRefreshTokenKey(tokenClaims.UserID, tokenClaims.ID)
	return authService.userCacheRepository.Delete(refreshTokenKey)
}

func (authService *AuthService) formatRefreshTokenKey(userID int64, tokenID string) string {
	return fmt.Sprintf("auth:refresh:%d:%s", userID, tokenID)
}
