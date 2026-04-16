package router

import "net/http"

func NewRouter(dependency RouterDependency) Router {
	return Router{
		healthHandler:    dependency.HealthHandler,
		methodMiddleware: dependency.MethodMiddleware,
	}
}

func (router Router) Handler() http.Handler {
	httpRouter := http.NewServeMux()

	router.registerHealthRoutes(httpRouter)
	router.registerAuthRoutes(httpRouter)
	router.registerThreadRoutes(httpRouter)

	return httpRouter
}
