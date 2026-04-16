package router

import (
	"net/http"

	"final-task/controller"
	"final-task/middleware"
)

type HTTPRouter struct {
	authController   controller.AuthControllerHandler
	threadController controller.ThreadControllerHandler
	verifyToken      middleware.VerifyTokenFunc
}

func NewHTTPRouter(
	authController controller.AuthControllerHandler,
	threadController controller.ThreadControllerHandler,
	verifyToken middleware.VerifyTokenFunc,
) HTTPRouter {
	return HTTPRouter{
		authController:   authController,
		threadController: threadController,
		verifyToken:      verifyToken,
	}
}

func (router HTTPRouter) Handler() http.Handler {
	httpRouter := http.NewServeMux()
	router.registerAuthRoutes(httpRouter)
	router.registerThreadRoutes(httpRouter)
	router.registerHealthRoutes(httpRouter)

	return httpRouter
}
