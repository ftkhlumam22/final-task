package health

import (
	"net/http"

	"kaktus-consumer/helper"
)

func NewController(dependency Dependency) Controller {
	return Controller{
		healthService: dependency.HealthService,
	}
}

func (controller Controller) Health() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, _ *http.Request) {
		helper.WriteSuccess(responseWriter, http.StatusOK, controller.healthService.Status())
	}
}
