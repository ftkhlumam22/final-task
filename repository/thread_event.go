package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"final-task/model"
	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishThreadCreatedEvent(
	requestContext context.Context,
	rabbitPublisher *model.RabbitPublisher,
	threadCreatedEvent model.ThreadCreatedEvent,
) error {
	return publishEvent(requestContext, rabbitPublisher, rabbitPublisher.ThreadCreatedRoutingKey, model.EventThreadCreated, threadCreatedEvent.RequestID, threadCreatedEvent)
}

func PublishCommentCreatedEvent(
	requestContext context.Context,
	rabbitPublisher *model.RabbitPublisher,
	commentCreatedEvent model.CommentCreatedEvent,
) error {
	return publishEvent(requestContext, rabbitPublisher, rabbitPublisher.CommentCreatedRoutingKey, model.EventCommentCreated, commentCreatedEvent.RequestID, commentCreatedEvent)
}

func PublishThreadLikedEvent(
	requestContext context.Context,
	rabbitPublisher *model.RabbitPublisher,
	threadLikedEvent model.ThreadLikedEvent,
) error {
	return publishEvent(requestContext, rabbitPublisher, rabbitPublisher.ThreadLikedRoutingKey, model.EventThreadLiked, threadLikedEvent.RequestID, threadLikedEvent)
}

func publishEvent(
	requestContext context.Context,
	rabbitPublisher *model.RabbitPublisher,
	routingKey string,
	eventName string,
	requestID string,
	eventPayloadData any,
) error {
	eventPayload, marshalError := json.Marshal(eventPayloadData)
	if marshalError != nil {
		return fmt.Errorf("%w: %v", model.ErrPublishEvent, marshalError)
	}

	log.Printf(
		"[PRODUCER] Mulai kirim event %s ke RabbitMQ. request_id=%s exchange=%s routing_key=%s",
		eventName,
		requestID,
		rabbitPublisher.ExchangeName,
		routingKey,
	)

	publishError := rabbitPublisher.Channel.PublishWithContext(
		requestContext,
		rabbitPublisher.ExchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         eventPayload,
		},
	)
	if publishError != nil {
		log.Printf("[PRODUCER] Gagal kirim event %s ke RabbitMQ. request_id=%s err=%v", eventName, requestID, publishError)
		return fmt.Errorf("%w: %v", model.ErrPublishEvent, publishError)
	}

	log.Printf("[PRODUCER] Event %s berhasil dikirim ke RabbitMQ. request_id=%s", eventName, requestID)
	return nil
}
