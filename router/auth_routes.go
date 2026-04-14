package router

import (
	"net/http"

	"final-task/helper"
)

func (router *HTTPRouter) registerAuthRoutes(httpRouter *http.ServeMux) {
	authRouteGroup := http.NewServeMux()
	authRouteGroup.HandleFunc("/register", helper.HandleMethod(http.MethodPost, router.authController.RegisterHandler))
	authRouteGroup.HandleFunc("/login", helper.HandleMethod(http.MethodPost, router.authController.LoginHandler))
	authRouteGroup.HandleFunc("/refresh", helper.HandleMethod(http.MethodPost, router.authController.RefreshHandler))
	authRouteGroup.HandleFunc("/logout", helper.HandleMethod(http.MethodPost, router.authController.LogoutHandler))
	httpRouter.Handle("/auth/", http.StripPrefix("/auth", authRouteGroup))
}
