package router

import "net/http"

func (router Router) registerHealthRoutes(httpRouter *http.ServeMux) {
	httpRouter.HandleFunc(
		"/health",
		router.methodMiddleware.Handle(http.MethodGet, router.healthHandler.Health()),
	)
}
