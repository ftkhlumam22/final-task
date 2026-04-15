package middleware

import "net/http"

type MethodMiddleware interface {
	Handle(expectedMethod string, nextHandler http.HandlerFunc) http.HandlerFunc
}

type methodMiddleware struct{}
