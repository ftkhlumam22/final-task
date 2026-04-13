package router

import (
	"database/sql"
	"net/http"

	"final-task/controller"
	"final-task/helper"
	"final-task/model"
	"github.com/redis/go-redis/v9"
)

func CollectRouter(
	databaseConnection *sql.DB,
	jwtManager *model.JWTManager,
	redisClient *redis.Client,
	rabbitPublisher *model.RabbitPublisher,
) http.Handler {
	httpRouter := http.NewServeMux()

	httpRouter.HandleFunc("/auth/register", helper.HandleMethod(http.MethodPost, controller.RegisterHandler(databaseConnection)))
	httpRouter.HandleFunc("/auth/login", helper.HandleMethod(http.MethodPost, controller.LoginHandler(databaseConnection, jwtManager, redisClient)))
	httpRouter.HandleFunc("/auth/refresh", helper.HandleMethod(http.MethodPost, controller.RefreshHandler(jwtManager, redisClient)))
	httpRouter.HandleFunc("/auth/logout", helper.HandleMethod(http.MethodPost, controller.LogoutHandler(jwtManager, redisClient)))
	httpRouter.HandleFunc("/threads", helper.HandleMethods(map[string]http.HandlerFunc{
		http.MethodGet:  controller.ListThreadHandler(databaseConnection, redisClient),
		http.MethodPost: helper.RequireAccessToken(jwtManager, controller.CreateThreadHandler(rabbitPublisher, redisClient)),
	}))
	httpRouter.HandleFunc("/threads/detail", helper.HandleMethod(http.MethodGet, controller.DetailThreadHandler(databaseConnection, redisClient)))
	httpRouter.HandleFunc("/threads/comment", helper.HandleMethod(http.MethodPost, helper.RequireAccessToken(jwtManager, controller.CreateCommentHandler(rabbitPublisher, redisClient))))
	httpRouter.HandleFunc("/threads/like", helper.HandleMethod(http.MethodPost, helper.RequireAccessToken(jwtManager, controller.InsertLikeThreadHandler(rabbitPublisher, redisClient))))

	httpRouter.HandleFunc("/health", helper.HandleMethod(http.MethodGet, func(responseWriter http.ResponseWriter, _ *http.Request) {
		responseWriter.WriteHeader(http.StatusOK)
		_, _ = responseWriter.Write([]byte("ok"))
	}))

	return httpRouter
}
