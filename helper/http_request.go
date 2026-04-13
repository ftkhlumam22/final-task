package helper

import (
	"encoding/json"
	"net/http"

	"final-task/model"
)

func DecodeRequestBody(httpRequest *http.Request, target any) error {
	requestDecoder := json.NewDecoder(httpRequest.Body)
	requestDecoder.DisallowUnknownFields()

	return requestDecoder.Decode(target)
}

func HandleMethod(expectedMethod string, nextHandler http.HandlerFunc) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		if httpRequest.Method != expectedMethod {
			http.Error(responseWriter, model.MessageMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}

		nextHandler(responseWriter, httpRequest)
	}
}

func HandleMethods(methodHandlers map[string]http.HandlerFunc) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		handler, exists := methodHandlers[httpRequest.Method]
		if !exists {
			http.Error(responseWriter, model.MessageMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}

		handler(responseWriter, httpRequest)
	}
}
