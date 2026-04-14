package middleware

import (
	"net/http"

	"kaktus-consumer/model"
)

func NewMethodMiddleware(dependency MethodMiddlewareDependency) MethodMiddleware {
	return &methodMiddleware{
		writeError: dependency.WriteError,
	}
}

func (middleware *methodMiddleware) Handle(
	expectedMethod string,
	nextHandler http.HandlerFunc,
) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method != expectedMethod {
			middleware.writeError(
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
