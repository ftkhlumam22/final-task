package controller

import (
	"database/sql"
	"net/http"
	"strings"

	"final-task/dto/request"
	"final-task/dto/response"
	"final-task/helper"
	"final-task/model"
	"final-task/module"

	"github.com/redis/go-redis/v9"
)

func RegisterHandler(databaseConnection *sql.DB) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		var registerRequest request.RegisterUser
		if decodeError := helper.DecodeRequestBody(httpRequest, &registerRequest); decodeError != nil {
			helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageInvalidRequestBody)
			return
		}

		registerRequest.Name = strings.TrimSpace(registerRequest.Name)
		registerRequest.Email = strings.TrimSpace(registerRequest.Email)
		registerRequest.Password = strings.TrimSpace(registerRequest.Password)
		if registerRequest.Name == "" || registerRequest.Email == "" || registerRequest.Password == "" {
			helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageNameEmailPasswordNeeded)
			return
		}

		createdUser, registerError := module.RegisterUser(
			httpRequest.Context(),
			databaseConnection,
			registerRequest.Name,
			registerRequest.Email,
			registerRequest.Password,
		)
		if registerError != nil {
			helper.WriteMappedError(responseWriter, registerError)
			return
		}

		helper.WriteSuccess(responseWriter, http.StatusCreated, response.RegisterUser{
			ID:        createdUser.ID,
			Email:     createdUser.Email,
			Name:      createdUser.Name,
			CreatedAt: createdUser.CreatedAt,
		})
	}
}

func LoginHandler(
	databaseConnection *sql.DB,
	jwtManager *model.JWTManager,
	redisClient *redis.Client,
) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		var loginRequest request.LoginUser
		if decodeError := helper.DecodeRequestBody(httpRequest, &loginRequest); decodeError != nil {
			helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageInvalidRequestBody)
			return
		}

		loginRequest.Email = strings.TrimSpace(loginRequest.Email)
		loginRequest.Password = strings.TrimSpace(loginRequest.Password)
		if loginRequest.Email == "" || loginRequest.Password == "" {
			helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageEmailPasswordNeeded)
			return
		}

		authResult, loginError := module.LoginUser(
			httpRequest.Context(),
			databaseConnection,
			jwtManager,
			redisClient,
			loginRequest.Email,
			loginRequest.Password,
		)
		if loginError != nil {
			helper.WriteMappedError(responseWriter, loginError)
			return
		}

		helper.WriteSuccess(responseWriter, http.StatusOK, response.LoginUser{
			Email:        authResult.User.Email,
			Name:         authResult.User.Name,
			AccessToken:  authResult.AccessToken,
			RefreshToken: authResult.RefreshToken,
		})
	}
}

func RefreshHandler(jwtManager *model.JWTManager, redisClient *redis.Client) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		var refreshRequest request.RefreshToken
		if decodeError := helper.DecodeRequestBody(httpRequest, &refreshRequest); decodeError != nil {
			helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageInvalidRequestBody)
			return
		}

		refreshRequest.RefreshToken = strings.TrimSpace(refreshRequest.RefreshToken)
		if refreshRequest.RefreshToken == "" {
			helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageRefreshTokenNeeded)
			return
		}

		authResult, refreshError := module.RefreshUserToken(
			httpRequest.Context(),
			jwtManager,
			redisClient,
			refreshRequest.RefreshToken,
		)
		if refreshError != nil {
			helper.WriteMappedError(responseWriter, refreshError)
			return
		}

		helper.WriteSuccess(responseWriter, http.StatusOK, map[string]string{
			"access_token":  authResult.AccessToken,
			"refresh_token": authResult.RefreshToken,
		})
	}
}

func LogoutHandler(jwtManager *model.JWTManager, redisClient *redis.Client) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		var refreshRequest request.RefreshToken
		if decodeError := helper.DecodeRequestBody(httpRequest, &refreshRequest); decodeError != nil {
			helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageInvalidRequestBody)
			return
		}

		refreshRequest.RefreshToken = strings.TrimSpace(refreshRequest.RefreshToken)
		if refreshRequest.RefreshToken == "" {
			helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageRefreshTokenNeeded)
			return
		}

		logoutError := module.LogoutUser(httpRequest.Context(), jwtManager, redisClient, refreshRequest.RefreshToken)
		if logoutError != nil {
			helper.WriteMappedError(responseWriter, logoutError)
			return
		}

		helper.WriteSuccess(responseWriter, http.StatusOK, map[string]string{"message": model.MessageLogoutSuccess})
	}
}
