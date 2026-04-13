package controller

import (
	"net/http"

	"kaktus-consumer/helper"
)

func HealthHandler() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, _ *http.Request) {
		helper.WriteSuccess(responseWriter, http.StatusOK, map[string]string{
			"status": "ok",
		})
	}
}
