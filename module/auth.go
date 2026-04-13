package module

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"final-task/helper"
	"final-task/model"
	"final-task/repository"
	"github.com/redis/go-redis/v9"
)

func RegisterUser(
	requestContext context.Context,
	databaseConnection *sql.DB,
	userName string,
	userEmail string,
	userPassword string,
) (model.User, error) {
	passwordHash, hashError := helper.HashPassword(userPassword)
	if hashError != nil {
		return model.User{}, hashError
	}

	return repository.CreateUser(requestContext, databaseConnection, userName, userEmail, passwordHash)
}

func LoginUser(
	requestContext context.Context,
	databaseConnection *sql.DB,
	jwtManager *model.JWTManager,
	redisClient *redis.Client,
	userEmail string,
	userPassword string,
) (model.AuthResult, error) {
	foundUser, findUserError := repository.FindUserByEmail(requestContext, databaseConnection, userEmail)
	if findUserError != nil {
		if errors.Is(findUserError, sql.ErrNoRows) {
			return model.AuthResult{}, model.ErrInvalidCredential
		}
		return model.AuthResult{}, findUserError
	}

	if passwordError := helper.ComparePassword(foundUser.PasswordHash, userPassword); passwordError != nil {
		return model.AuthResult{}, model.ErrInvalidCredential
	}

	accessToken, accessTokenError := jwtManager.GenerateAccessToken(foundUser.ID)
	if accessTokenError != nil {
		return model.AuthResult{}, accessTokenError
	}

	refreshToken, refreshTokenID, refreshTokenError := jwtManager.GenerateRefreshToken(foundUser.ID)
	if refreshTokenError != nil {
		return model.AuthResult{}, refreshTokenError
	}

	refreshTokenKey := formatRefreshTokenKey(foundUser.ID, refreshTokenID)
	redisError := helper.RedisSetWithTTL(requestContext, redisClient, refreshTokenKey, "1", jwtManager.RefreshTTL())
	if redisError != nil {
		return model.AuthResult{}, redisError
	}

	return model.AuthResult{
		User:         foundUser,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func RefreshUserToken(
	requestContext context.Context,
	jwtManager *model.JWTManager,
	redisClient *redis.Client,
	refreshToken string,
) (model.AuthResult, error) {
	tokenClaims, tokenError := jwtManager.VerifyToken(refreshToken, model.TokenTypeRefresh)
	if tokenError != nil {
		return model.AuthResult{}, model.ErrInvalidRefresh
	}

	refreshTokenKey := formatRefreshTokenKey(tokenClaims.UserID, tokenClaims.ID)
	refreshTokenExists, redisExistsError := helper.RedisExists(requestContext, redisClient, refreshTokenKey)
	if redisExistsError != nil {
		return model.AuthResult{}, redisExistsError
	}
	if !refreshTokenExists {
		return model.AuthResult{}, model.ErrInvalidRefresh
	}

	redisDeleteError := helper.RedisDelete(requestContext, redisClient, refreshTokenKey)
	if redisDeleteError != nil {
		return model.AuthResult{}, redisDeleteError
	}

	accessToken, accessTokenError := jwtManager.GenerateAccessToken(tokenClaims.UserID)
	if accessTokenError != nil {
		return model.AuthResult{}, accessTokenError
	}

	newRefreshToken, newRefreshTokenID, newRefreshTokenError := jwtManager.GenerateRefreshToken(tokenClaims.UserID)
	if newRefreshTokenError != nil {
		return model.AuthResult{}, newRefreshTokenError
	}

	newRefreshTokenKey := formatRefreshTokenKey(tokenClaims.UserID, newRefreshTokenID)
	redisSetError := helper.RedisSetWithTTL(requestContext, redisClient, newRefreshTokenKey, "1", jwtManager.RefreshTTL())
	if redisSetError != nil {
		return model.AuthResult{}, redisSetError
	}

	return model.AuthResult{
		User: model.User{
			ID: tokenClaims.UserID,
		},
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func LogoutUser(
	requestContext context.Context,
	jwtManager *model.JWTManager,
	redisClient *redis.Client,
	refreshToken string,
) error {
	tokenClaims, tokenError := jwtManager.VerifyToken(refreshToken, model.TokenTypeRefresh)
	if tokenError != nil {
		return model.ErrInvalidRefresh
	}

	refreshTokenKey := formatRefreshTokenKey(tokenClaims.UserID, tokenClaims.ID)
	return helper.RedisDelete(requestContext, redisClient, refreshTokenKey)
}

func formatRefreshTokenKey(userID int64, tokenID string) string {
	return fmt.Sprintf("auth:refresh:%d:%s", userID, tokenID)
}
