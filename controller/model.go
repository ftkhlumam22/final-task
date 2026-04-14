package controller

import (
	"context"
	"net/http"

	"kaktus-consumer/module"
	"kaktus-consumer/repository"
)

const (
	threadCreatedConsumerTag = "kaktus-consumer-thread-created"
	commentCreatedConsumerTag = "kaktus-consumer-comment-created"
	threadLikedConsumerTag   = "kaktus-consumer-thread-liked"
	threadGetLikedConsumerTag = "kaktus-consumer-thread-get-liked"
)

type EventConsumerHandler interface {
	Start(requestContext context.Context) error
}

type HealthHandler interface {
	Health() http.HandlerFunc
}

type EventConsumerControllerDependency struct {
	ConsumerService      module.ConsumerService
	SubscriberRepository repository.EventSubscriberRepository
	ThreadCreatedQueueName string
	CommentCreatedQueueName string
	ThreadLikedQueueName string
	ThreadGetLikedQueueName string
}

type HealthControllerDependency struct {
	HealthService module.HealthService
}

type EventConsumerController struct {
	consumerService      module.ConsumerService
	subscriberRepository repository.EventSubscriberRepository
	threadCreatedQueueName string
	commentCreatedQueueName string
	threadLikedQueueName string
	threadGetLikedQueueName string
}

type HealthController struct {
	healthService module.HealthService
}
