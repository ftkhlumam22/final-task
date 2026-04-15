package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"kaktus-consumer/config"
	"kaktus-consumer/controller"
	"kaktus-consumer/messaging"
	"kaktus-consumer/messaging/publisher"
	"kaktus-consumer/messaging/subscriber"
	"kaktus-consumer/middleware"
	"kaktus-consumer/model"
	"kaktus-consumer/module"
	"kaktus-consumer/repository"
	"kaktus-consumer/router"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("warning: failed to load .env file: %v", err)
	}

	environmentFactory := model.NewOSEnvFactory()
	applicationConfig, err := config.LoadAppConfig(environmentFactory)
	if err != nil {
		log.Fatal(err)
	}

	databaseConnection, err := config.NewDB(applicationConfig.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer databaseConnection.Close()

	redisClient, err := config.NewRedis(
		applicationConfig.RedisURL,
		applicationConfig.RedisPassword,
		applicationConfig.RedisDB,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()

	rabbitConnection, err := config.NewRabbitConnection(applicationConfig.RabbitMQ.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitConnection.Close()

	rabbitSubscriberChannel, err := rabbitConnection.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitSubscriberChannel.Close()

	rabbitTopologyConfig := messaging.LoadRabbitTopologyConfig(environmentFactory)
	consumerQueues, err := messaging.SetupConsumerTopology(rabbitSubscriberChannel, rabbitTopologyConfig)
	if err != nil {
		log.Fatal(err)
	}

	err = rabbitSubscriberChannel.Qos(10, 0, false)
	if err != nil {
		log.Fatal(err)
	}

	subscriberRepository := subscriber.NewSubscriberRepository(rabbitSubscriberChannel)

	publisherRepository, err := publisher.NewPublisherRepository(rabbitConnection)
	if err != nil {
		log.Fatal(err)
	}
	defer publisherRepository.Close()

	threadRepository := repository.NewThreadRepository(repository.ThreadRepositoryDependency{
		DatabaseConnection: databaseConnection,
		RedisClient:        redisClient,
	})

	threadService := module.NewThreadService(module.ThreadServiceDependency{
		ThreadRepository:    threadRepository,
		PublisherRepository: publisherRepository,
		MaxRetry:            applicationConfig.RabbitMQ.MaxRetry,
		RetryDelay:          time.Duration(model.ConsumerRetryDelaySeconds) * time.Second,
	})

	healthService := module.NewHealthService()

	requestContext, cancelContext := context.WithCancel(context.Background())
	defer cancelContext()

	threadConsumerController := controller.NewThreadConsumerController(
		controller.ThreadConsumerControllerDependency{
			ThreadService:           threadService,
			SubscriberRepository:    subscriberRepository,
			ThreadCreatedQueueName:  consumerQueues.ThreadCreatedQueueName,
			CommentCreatedQueueName: consumerQueues.CommentCreatedQueueName,
			ThreadLikedQueueName:    consumerQueues.ThreadLikedQueueName,
			ThreadGetLikedQueueName: consumerQueues.ThreadGetLikedQueueName,
		},
	)

	go func() {
		err := threadConsumerController.Start(requestContext)
		if err != nil {
			log.Fatal(err)
		}
	}()

	healthController := controller.NewHealthController(controller.HealthControllerDependency{
		HealthService: healthService,
	})

	methodMiddleware := middleware.NewMethodMiddleware()

	httpRouter := router.NewRouter(router.RouterDependency{
		HealthHandler:    healthController,
		MethodMiddleware: methodMiddleware,
	})

	httpServer := http.Server{
		Addr:         applicationConfig.ServerAddr,
		Handler:      httpRouter.Handler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("consumer running %s", applicationConfig.ServerAddr)
	err = httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
