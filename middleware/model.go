package middleware

import "net/http"

type ErrorWriter func(responseWriter http.ResponseWriter, statusCode int, message string, errorDetail interface{})

type MethodMiddleware interface {
	Handle(expectedMethod string, nextHandler http.HandlerFunc) http.HandlerFunc
}

type MethodMiddlewareDependency struct {
	WriteError ErrorWriter
}

type methodMiddleware struct {
	writeError ErrorWriter
}
