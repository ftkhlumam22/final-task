package health

import (
	"net/http"

	"kaktus-consumer/module"
)

type Handler interface {
	Health() http.HandlerFunc
}

type Dependency struct {
	HealthService module.HealthService
}

type Controller struct {
	healthService module.HealthService
}
