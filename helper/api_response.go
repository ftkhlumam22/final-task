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

func WriteMappedError(responseWriter http.ResponseWriter, err error) {
	statusCode, message := mapError(err)
	WriteError(responseWriter, statusCode, message)
}

func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, model.ErrEmailAlreadyExists):
		return http.StatusConflict, model.MessageEmailAlreadyExists
	case errors.Is(err, model.ErrInvalidCredential):
		return http.StatusUnauthorized, model.MessageInvalidCredential
	case errors.Is(err, model.ErrInvalidRefresh):
		return http.StatusUnauthorized, model.MessageInvalidRefresh
	case errors.Is(err, model.ErrPublishEvent):
		return http.StatusBadGateway, model.MessageFailedPublishEvent
	case errors.Is(err, model.ErrThreadNotFound):
		return http.StatusNotFound, model.MessageThreadNotFound
	case errors.Is(err, model.ErrConsumeEvent):
		return http.StatusBadGateway, model.MessageFailedConsumeEvent
	default:
		return http.StatusInternalServerError, model.MessageInternalServerError
	}
}
