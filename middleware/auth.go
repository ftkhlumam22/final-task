package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"final-task/helper"
	"final-task/model"
)

type VerifyTokenFunc func(tokenString string, expectedTokenType string) (model.CustomClaims, error)

func RequireAccessToken(verifyToken VerifyTokenFunc, nextHandler http.HandlerFunc) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		authorizationHeader := strings.TrimSpace(httpRequest.Header.Get("Authorization"))
		headerParts := strings.SplitN(authorizationHeader, " ", 2)
		if len(headerParts) != 2 || !strings.EqualFold(headerParts[0], "Bearer") {
			log.Printf("[AUTH] Header Authorization tidak valid.")
			helper.WriteError(responseWriter, http.StatusUnauthorized, model.MessageUnauthorized)
			return
		}

		accessToken := strings.TrimSpace(headerParts[1])
		if accessToken == "" {
			log.Printf("[AUTH] Access token kosong.")
			helper.WriteError(responseWriter, http.StatusUnauthorized, model.MessageUnauthorized)
			return
		}

		tokenClaims, err := verifyToken(accessToken, model.TokenTypeAccess)
		if err != nil {
			log.Printf("[AUTH] Access token tidak valid. err=%v", err)
			helper.WriteError(responseWriter, http.StatusUnauthorized, model.MessageUnauthorized)
			return
		}
		log.Printf("[AUTH] Access token valid. user_id=%d", tokenClaims.UserID)

		requestContext := context.WithValue(httpRequest.Context(), model.ContextKeyUserID, tokenClaims.UserID)
		nextHandler(responseWriter, httpRequest.WithContext(requestContext))
	}
}

func UserIDFromRequest(httpRequest *http.Request) (int64, error) {
	userID, isValidType := httpRequest.Context().Value(model.ContextKeyUserID).(int64)
	if !isValidType || userID <= 0 {
		return 0, errors.New(model.MessageUnauthorized)
	}

	return userID, nil
}
