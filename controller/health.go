package controller

import (
	"net/http"

	"kaktus-consumer/helper"
)

func NewHealthController(dependency HealthControllerDependency) *HealthController {
	return &HealthController{
		healthService: dependency.HealthService,
	}
}

func (controller *HealthController) Health() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, _ *http.Request) {
		helper.WriteSuccess(responseWriter, http.StatusOK, controller.healthService.Status())
	}
}
