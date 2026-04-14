package router

import (
	"kaktus-consumer/controller"
	"kaktus-consumer/middleware"
)

type RouterDependency struct {
	HealthHandler   controller.HealthHandler
	MethodMiddleware middleware.MethodMiddleware
}

type Router struct {
	healthHandler    controller.HealthHandler
	methodMiddleware middleware.MethodMiddleware
}
