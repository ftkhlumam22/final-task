package module

import (
	"context"
	"time"

	"kaktus-consumer/model"
	"kaktus-consumer/repository"
)

type EventService interface {
	ParseEventEnvelope(messageBody []byte) (model.EventEnvelope, error)
	ProcessEvent(requestContext context.Context, eventName string, messageBody []byte) error
	BuildThreadGetLikedResponse(requestContext context.Context, messageBody []byte) (interface{}, error)
}

type HealthService interface {
	Status() map[string]string
}

type ConsumerService interface {
	HandleThreadCreatedMessage(requestContext context.Context, messageBody []byte) ConsumeDecision
	HandleCommentCreatedMessage(requestContext context.Context, messageBody []byte) ConsumeDecision
	HandleThreadLikedMessage(requestContext context.Context, messageBody []byte) ConsumeDecision
	HandleThreadGetLikedMessage(
		requestContext context.Context,
		messageBody []byte,
		replyToQueue string,
		correlationID string,
	) ConsumeDecision
}

type ConsumeDecision struct {
	EventName  string
	RequestID  string
	ShouldAck  bool
	Requeue    bool
	NackReason string
	Err        error
}

type EventServiceDependency struct {
	SQLRepository   repository.SQLRepository
	CacheRepository repository.CacheRepository
}

type ConsumerServiceDependency struct {
	EventService         EventService
	PublisherRepository  repository.RPCPublisherRepository
	MaxRetry             int
	RetryDelay           time.Duration
}

type eventService struct {
	sqlRepository   repository.SQLRepository
	cacheRepository repository.CacheRepository
}

type consumerService struct {
	eventService        EventService
	publisherRepository repository.RPCPublisherRepository
	maxRetry            int
	retryDelay          time.Duration
}

type healthService struct{}
