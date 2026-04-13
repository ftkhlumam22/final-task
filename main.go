package main

import (
	"final-task/config"
	"final-task/model"
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

	rabbitPublisher, err := config.NewRabbitPublisher(appConfig.RabbitMQ)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitPublisher.Channel.Close()
	defer rabbitPublisher.Connection.Close()
	log.Printf("[BOOT] Koneksi RabbitMQ berhasil. exchange=%s", rabbitPublisher.ExchangeName)

	jwtManager := model.NewJWTManager(appConfig.JWT)
	log.Printf("[BOOT] JWT manager siap dipakai.")

	httpServer := http.Server{
		Addr:         appConfig.ServerAddr,
		Handler:      router.CollectRouter(databaseConnection, jwtManager, redisClient, rabbitPublisher),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("running %s", appConfig.ServerAddr)

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("start server: %v", err)
	}
}
