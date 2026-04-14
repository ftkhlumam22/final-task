package router

import (
	"net/http"

	"final-task/helper"
)

func (router *HTTPRouter) registerHealthRoutes(httpRouter *http.ServeMux) {
	httpRouter.HandleFunc("/health", helper.HandleMethod(http.MethodGet, func(responseWriter http.ResponseWriter, _ *http.Request) {
		responseWriter.WriteHeader(http.StatusOK)
		_, _ = responseWriter.Write([]byte("ok"))
	}))
}
