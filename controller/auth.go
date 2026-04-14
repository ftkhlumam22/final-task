package controller

import (
	"net/http"
	"strings"

	"final-task/dto/request"
	"final-task/dto/response"
	"final-task/helper"
	"final-task/model"
)

func NewAuthController(authService AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (authController *AuthController) RegisterHandler(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	var registerRequest request.RegisterUser
	if err := helper.DecodeRequestBody(httpRequest, &registerRequest); err != nil {
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

	createdUser, err := authController.authService.RegisterUser(
		registerRequest.Name,
		registerRequest.Email,
		registerRequest.Password,
	)
	if err != nil {
		helper.WriteMappedError(responseWriter, err)
		return
	}

	helper.WriteSuccess(responseWriter, http.StatusCreated, response.RegisterUser{
		ID:        createdUser.ID,
		Email:     createdUser.Email,
		Name:      createdUser.Name,
		CreatedAt: createdUser.CreatedAt,
	})
}

func (authController *AuthController) LoginHandler(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	var loginRequest request.LoginUser
	if err := helper.DecodeRequestBody(httpRequest, &loginRequest); err != nil {
		helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageInvalidRequestBody)
		return
	}

	loginRequest.Email = strings.TrimSpace(loginRequest.Email)
	loginRequest.Password = strings.TrimSpace(loginRequest.Password)
	if loginRequest.Email == "" || loginRequest.Password == "" {
		helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageEmailPasswordNeeded)
		return
	}

	authResult, err := authController.authService.LoginUser(
		loginRequest.Email,
		loginRequest.Password,
	)
	if err != nil {
		helper.WriteMappedError(responseWriter, err)
		return
	}

	helper.WriteSuccess(responseWriter, http.StatusOK, response.LoginUser{
		Email:        authResult.User.Email,
		Name:         authResult.User.Name,
		AccessToken:  authResult.AccessToken,
		RefreshToken: authResult.RefreshToken,
	})
}

func (authController *AuthController) RefreshHandler(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	var refreshRequest request.RefreshToken
	if err := helper.DecodeRequestBody(httpRequest, &refreshRequest); err != nil {
		helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageInvalidRequestBody)
		return
	}

	refreshRequest.RefreshToken = strings.TrimSpace(refreshRequest.RefreshToken)
	if refreshRequest.RefreshToken == "" {
		helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageRefreshTokenNeeded)
		return
	}

	authResult, err := authController.authService.RefreshUserToken(
		refreshRequest.RefreshToken,
	)
	if err != nil {
		helper.WriteMappedError(responseWriter, err)
		return
	}

	helper.WriteSuccess(responseWriter, http.StatusOK, map[string]string{
		"access_token":  authResult.AccessToken,
		"refresh_token": authResult.RefreshToken,
	})
}

func (authController *AuthController) LogoutHandler(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	var refreshRequest request.RefreshToken
	if err := helper.DecodeRequestBody(httpRequest, &refreshRequest); err != nil {
		helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageInvalidRequestBody)
		return
	}

	refreshRequest.RefreshToken = strings.TrimSpace(refreshRequest.RefreshToken)
	if refreshRequest.RefreshToken == "" {
		helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageRefreshTokenNeeded)
		return
	}

	err := authController.authService.LogoutUser(refreshRequest.RefreshToken)
	if err != nil {
		helper.WriteMappedError(responseWriter, err)
		return
	}

	helper.WriteSuccess(responseWriter, http.StatusOK, map[string]string{"message": model.MessageLogoutSuccess})
}
