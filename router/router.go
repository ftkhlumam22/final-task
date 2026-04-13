package router

import (
	"net/http"

	"kaktus-consumer/controller"
	"kaktus-consumer/helper"
)

func CollectRouter() http.Handler {
	httpRouter := http.NewServeMux()

	httpRouter.HandleFunc("/health", helper.HandleMethod(http.MethodGet, controller.HealthHandler()))

	return httpRouter
}
