package helper

import (
	"encoding/json"
	"errors"
	"net/http"

	"kaktus-consumer/dto/response"
	"kaktus-consumer/model"
)

func WriteSuccess(responseWriter http.ResponseWriter, statusCode int, data interface{}) {
	writeJSON(responseWriter, statusCode, response.GeneralResponse{
		Status: model.StatusSuccess,
		Data:   data,
	})
}

func WriteError(responseWriter http.ResponseWriter, statusCode int, message string, errorDetail interface{}) {
	writeJSON(responseWriter, statusCode, response.GeneralResponse{
		Status:  model.StatusError,
		Message: message,
		Error:   errorDetail,
	})
}

func WriteMappedError(responseWriter http.ResponseWriter, err error) {
	if errors.Is(err, model.ErrInvalidEventPayload) {
		WriteError(responseWriter, http.StatusBadRequest, model.MessageInvalidEventPayload, nil)
		return
	}

	WriteError(responseWriter, http.StatusInternalServerError, model.MessageInternalServerError, nil)
}

func writeJSON(responseWriter http.ResponseWriter, statusCode int, responseBody response.GeneralResponse) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)

	if err := json.NewEncoder(responseWriter).Encode(responseBody); err != nil {
		http.Error(responseWriter, model.MessageInternalServerError, http.StatusInternalServerError)
	}
}
