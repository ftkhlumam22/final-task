package helper

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"kaktus-consumer/model"
)

func DecodeRequestBody(request *http.Request, destination interface{}) error {
	jsonDecoder := json.NewDecoder(request.Body)
	jsonDecoder.DisallowUnknownFields()

	if decodeError := jsonDecoder.Decode(destination); decodeError != nil {
		return decodeError
	}

	var trailingPayload map[string]interface{}
	if trailingPayloadError := jsonDecoder.Decode(&trailingPayload); !errors.Is(trailingPayloadError, io.EOF) {
		return errors.New("request body must only contain a single json object")
	}

	return nil
}

func HandleMethod(expectedMethod string, nextHandler http.HandlerFunc) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method != expectedMethod {
			WriteError(responseWriter, http.StatusMethodNotAllowed, model.MessageMethodNotAllowed, nil)
			return
		}

		nextHandler(responseWriter, request)
	}
}
