package middleware

import (
	"net/http"

	"kaktus-consumer/helper"
	"kaktus-consumer/model"
)

func NewMethodMiddleware() MethodMiddleware {
	return methodMiddleware{}
}

func (middleware methodMiddleware) Handle(
	expectedMethod string,
	nextHandler http.HandlerFunc,
) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method != expectedMethod {
			helper.WriteError(
				responseWriter,
				http.StatusMethodNotAllowed,
				model.MessageMethodNotAllowed,
				nil,
			)
			return
		}

		nextHandler(responseWriter, request)
	}
}
