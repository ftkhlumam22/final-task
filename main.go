package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"kaktus-consumer/config"
	"kaktus-consumer/controller"
	"kaktus-consumer/model"
	"kaktus-consumer/router"

	"github.com/joho/godotenv"
)

func main() {
	if environmentError := godotenv.Load(".env"); environmentError != nil {
		log.Printf("warning: failed to load .env file: %v", environmentError)
	}

	environmentFactory := model.NewOSEnvFactory()
	applicationConfig, configurationError := config.LoadAppConfig(environmentFactory)
	if configurationError != nil {
		log.Fatal(configurationError)
	}

	databaseConnection, databaseError := config.NewDB(applicationConfig.DatabaseURL)
	if databaseError != nil {
		log.Fatal(databaseError)
	}
	defer databaseConnection.Close()

	redisClient, redisError := config.NewRedis(
		applicationConfig.RedisURL,
		applicationConfig.RedisPassword,
		applicationConfig.RedisDB,
	)
	if redisError != nil {
		log.Fatal(redisError)
	}
	defer redisClient.Close()

	rabbitConsumer, rabbitError := config.NewRabbitConsumer(applicationConfig.RabbitMQ)
	if rabbitError != nil {
		log.Fatal(rabbitError)
	}
	defer rabbitConsumer.Channel.Close()
	defer rabbitConsumer.Connection.Close()

	requestContext, cancelContext := context.WithCancel(context.Background())
	defer cancelContext()

	go func() {
		workerError := controller.StartEventConsumerWorker(
			requestContext,
			rabbitConsumer,
			databaseConnection,
			redisClient,
		)
		if workerError != nil {
			log.Fatal(workerError)
		}
	}()

	httpServer := http.Server{
		Addr:         applicationConfig.ServerAddr,
		Handler:      router.CollectRouter(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("consumer running %s", applicationConfig.ServerAddr)
	if runError := httpServer.ListenAndServe(); runError != nil && runError != http.ErrServerClosed {
		log.Fatal(runError)
	}
}
