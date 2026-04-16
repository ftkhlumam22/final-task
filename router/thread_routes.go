package router

import (
	"net/http"

	"final-task/helper"
	"final-task/middleware"
)

func (router HTTPRouter) registerThreadRoutes(httpRouter *http.ServeMux) {
	httpRouter.HandleFunc("/threads", helper.HandleMethods(map[string]http.HandlerFunc{
		http.MethodGet:  router.threadController.ListThreadHandler,
		http.MethodPost: middleware.RequireAccessToken(router.verifyToken, router.threadController.CreateThreadHandler),
	}))

	threadRouteGroup := http.NewServeMux()
	threadRouteGroup.HandleFunc("/detail", helper.HandleMethod(http.MethodGet, router.threadController.DetailThreadHandler))
	threadRouteGroup.HandleFunc("/liked", helper.HandleMethod(http.MethodGet, router.threadController.GetLikedThread))
	threadRouteGroup.HandleFunc("/comment", helper.HandleMethod(http.MethodPost, middleware.RequireAccessToken(router.verifyToken, router.threadController.CreateCommentHandler)))
	threadRouteGroup.HandleFunc("/like", helper.HandleMethod(http.MethodPost, middleware.RequireAccessToken(router.verifyToken, router.threadController.InsertLikeThreadHandler)))
	httpRouter.Handle("/threads/", http.StripPrefix("/threads", threadRouteGroup))
}
