package thread

import (
	"context"

	"kaktus-consumer/module"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	threadCreatedConsumerTag  = "kaktus-consumer-thread-created"
	commentCreatedConsumerTag = "kaktus-consumer-comment-created"
	threadLikedConsumerTag    = "kaktus-consumer-thread-liked"
	threadGetLikedConsumerTag = "kaktus-consumer-thread-get-liked"
)

type SubscriberRepository interface {
	Consume(queueName string, consumerTag string) (<-chan amqp.Delivery, error)
	Ack(deliveryMessage amqp.Delivery) error
	Nack(deliveryMessage amqp.Delivery, requeue bool) error
}

type Handler interface {
	Start(requestContext context.Context) error
}

type Dependency struct {
	ThreadService           module.ThreadService
	SubscriberRepository    SubscriberRepository
	ThreadCreatedQueueName  string
	CommentCreatedQueueName string
	ThreadLikedQueueName    string
	ThreadGetLikedQueueName string
}

type Controller struct {
	threadService           module.ThreadService
	subscriberRepository    SubscriberRepository
	threadCreatedQueueName  string
	commentCreatedQueueName string
	threadLikedQueueName    string
	threadGetLikedQueueName string
}

func invalidPayloadDecision(eventName string, requestID string, err error) module.ConsumeDecision {
	return module.ConsumeDecision{
		EventName:  eventName,
		RequestID:  requestID,
		ShouldAck:  false,
		Requeue:    false,
		NackReason: "payload invalid",
		Err:        err,
	}
}
