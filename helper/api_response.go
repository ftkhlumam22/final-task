package helper

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"final-task/dto/response"
	"final-task/model"
)

func WriteSuccess(responseWriter http.ResponseWriter, statusCode int, payload any) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)

	_ = json.NewEncoder(responseWriter).Encode(response.GeneralSuccess{
		Status:     model.StatusSuccess,
		Data:       payload,
		Code:       statusCode,
		AccessTime: time.Now(),
	})
}

func WriteError(responseWriter http.ResponseWriter, statusCode int, message string) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)

	_ = json.NewEncoder(responseWriter).Encode(response.GeneralError{
		Status:     model.StatusError,
		Message:    message,
		Code:       statusCode,
		AccessTime: time.Now(),
	})
}

func WriteMappedError(responseWriter http.ResponseWriter, mappedError error) {
	statusCode, message := mapError(mappedError)
	WriteError(responseWriter, statusCode, message)
}

func mapError(mappedError error) (int, string) {
	switch {
	case errors.Is(mappedError, model.ErrEmailAlreadyExists):
		return http.StatusConflict, model.MessageEmailAlreadyExists
	case errors.Is(mappedError, model.ErrInvalidCredential):
		return http.StatusUnauthorized, model.MessageInvalidCredential
	case errors.Is(mappedError, model.ErrInvalidRefresh):
		return http.StatusUnauthorized, model.MessageInvalidRefresh
	case errors.Is(mappedError, model.ErrPublishEvent):
		return http.StatusBadGateway, model.MessageFailedPublishEvent
	case errors.Is(mappedError, model.ErrThreadNotFound):
		return http.StatusNotFound, model.MessageThreadNotFound
	default:
		return http.StatusInternalServerError, model.MessageInternalServerError
	}
}
