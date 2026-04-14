package repository

import (
	"context"

	"kaktus-consumer/model"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SQLRepository interface {
	InsertThread(requestContext context.Context, threadCreatedEvent model.ThreadCreatedEvent) (model.Thread, error)
	WithTransaction(requestContext context.Context, operation func(sqlTransactionRepository SQLTransactionRepository) error) error
	GetLikedThreads(requestContext context.Context) ([]model.ThreadLikedListItem, error)
}

type SQLTransactionRepository interface {
	InsertComment(requestContext context.Context, commentCreatedEvent model.CommentCreatedEvent) error
	IncrementParentCommentReply(requestContext context.Context, parentCommentID int64, threadID int64) (int64, error)
	IncrementThreadTotalComment(requestContext context.Context, threadID int64) (int64, error)
	InsertThreadLike(requestContext context.Context, threadLikedEvent model.ThreadLikedEvent) (int64, error)
	IncrementThreadTotalLike(requestContext context.Context, threadID int64) (int64, error)
}

type CacheRepository interface {
	ScanKeys(requestContext context.Context, keyPattern string, batchSize int64) ([]string, error)
	Delete(requestContext context.Context, keys ...string) error
}

type EventSubscriberRepository interface {
	Consume(queueName string, consumerTag string) (<-chan amqp.Delivery, error)
	Ack(deliveryMessage amqp.Delivery) error
	Nack(deliveryMessage amqp.Delivery, requeue bool) error
}

type RPCPublisherRepository interface {
	PublishRPCResponse(
		requestContext context.Context,
		replyToQueue string,
		correlationID string,
		responsePayload interface{},
	) error
	Close() error
}

type Dependency struct {
	SQL             SQLRepository
	Cache           CacheRepository
	EventSubscriber EventSubscriberRepository
	RPCPublisher    RPCPublisherRepository
}
