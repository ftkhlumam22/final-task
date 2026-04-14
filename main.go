package main

import (
	"final-task/config"
	"final-task/controller"
	"final-task/helper"
	"final-task/messaging"
	"final-task/model"
	authmodule "final-task/module/auth"
	threadmodule "final-task/module/thread"
	"final-task/repository/cacherepo"
	"final-task/repository/publisherrepo"
	"final-task/repository/sqlrepo"
	"final-task/router"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")

	environmentFactory := model.NewOSEnvFactory()
	appConfig, err := config.LoadAppConfig(environmentFactory)
	if err != nil {
		log.Fatal(err)
	}

	databaseConnection, err := config.NewDB(appConfig.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer databaseConnection.Close()
	log.Printf("[BOOT] Koneksi database berhasil.")

	redisClient, err := config.NewRedis(appConfig.RedisURL, appConfig.RedisPassword, appConfig.RedisDB)
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()
	log.Printf("[BOOT] Koneksi redis berhasil.")

	rabbitPublisher, err := messaging.NewRabbitPublisher(appConfig.RabbitMQ)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitPublisher.Channel.Close()
	defer rabbitPublisher.Connection.Close()
	log.Printf("[BOOT] Koneksi RabbitMQ berhasil. exchange=%s", rabbitPublisher.ExchangeName)

	rabbitRPCClient, err := messaging.NewRabbitRPCClient(appConfig.RabbitMQ)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitRPCClient.Channel.Close()
	defer rabbitRPCClient.Connection.Close()
	log.Printf("[BOOT] Koneksi RabbitMQ RPC Client berhasil. exchange=%s", rabbitRPCClient.ExchangeName)

	jwtManager := helper.NewJWTManager(appConfig.JWT)
	log.Printf("[BOOT] JWT manager siap dipakai.")
	tokenProvider := authmodule.NewJWTTokenProvider(jwtManager)

	userRepository := sqlrepo.NewUserRepository(databaseConnection)
	threadReadRepository := sqlrepo.NewThreadReadRepository(databaseConnection)
	redisRepository := cacherepo.NewRedisRepository(redisClient)
	userCacheRepository := cacherepo.NewUserCacheRepository(redisRepository)
	threadListCacheRepository := cacherepo.NewThreadListCacheRepository(redisRepository)
	threadDetailCacheRepository := cacherepo.NewThreadDetailCacheRepository(redisRepository)
	threadEventPublisher := publisherrepo.NewThreadEventPublisher(rabbitPublisher)
	threadRPCPublisher := publisherrepo.NewThreadRPCPublisher(rabbitRPCClient)

	authService := authmodule.NewAuthService(userRepository, tokenProvider, userCacheRepository)
	threadService := threadmodule.NewThreadService(
		threadReadRepository,
		threadListCacheRepository,
		threadDetailCacheRepository,
		threadEventPublisher,
		threadRPCPublisher,
	)

	authController := controller.NewAuthController(authService)
	threadController := controller.NewThreadController(threadService)
	httpRouter := router.NewHTTPRouter(authController, threadController, tokenProvider.VerifyToken)

	httpServer := http.Server{
		Addr:         appConfig.ServerAddr,
		Handler:      httpRouter.Handler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("running %s", appConfig.ServerAddr)

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("start server: %v", err)
	}
}
