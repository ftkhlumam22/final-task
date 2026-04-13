package controller

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"kaktus-consumer/model"
	"kaktus-consumer/module"

	"github.com/redis/go-redis/v9"
	amqp "github.com/rabbitmq/amqp091-go"
)

func StartEventConsumerWorker(
	requestContext context.Context,
	rabbitConsumer *model.RabbitConsumer,
	databaseConnection *sql.DB,
	redisClient *redis.Client,
) error {
	const consumerTag = "kaktus-consumer-events"

	deliveryChannel, consumeError := rabbitConsumer.Channel.Consume(
		rabbitConsumer.QueueName,
		consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if consumeError != nil {
		return fmt.Errorf("%w: start consume: %v", model.ErrConsumeEvent, consumeError)
	}

	log.Printf("[consumer] worker aktif. queue=%s max_retry=%d", rabbitConsumer.QueueName, rabbitConsumer.MaxRetry)

	for {
		select {
		case <-requestContext.Done():
			log.Printf("[consumer] worker berhenti: context selesai")
			return nil
		case deliveryMessage, isChannelOpen := <-deliveryChannel:
			if !isChannelOpen {
				return fmt.Errorf("%w: rabbit delivery channel closed", model.ErrConsumeEvent)
			}
			processDelivery(requestContext, rabbitConsumer.MaxRetry, databaseConnection, redisClient, deliveryMessage)
		}
	}
}

func processDelivery(
	requestContext context.Context,
	maxRetry int,
	databaseConnection *sql.DB,
	redisClient *redis.Client,
	deliveryMessage amqp.Delivery,
) {
	eventEnvelope, parseEnvelopeError := module.ParseEventEnvelope(deliveryMessage.Body)
	if parseEnvelopeError != nil {
		log.Printf("[consumer] validasi payload gagal: envelope tidak valid")
		nackMessage(deliveryMessage, "payload invalid", false)
		return
	}

	log.Printf("[consumer] pesan diterima. event=%s request_id=%s", eventEnvelope.Event, eventEnvelope.RequestID)

	var processingError error
	for retryAttempt := 0; retryAttempt <= maxRetry; retryAttempt++ {
		if retryAttempt > 0 {
			log.Printf("[consumer] retry ke-%d. event=%s request_id=%s", retryAttempt, eventEnvelope.Event, eventEnvelope.RequestID)
			time.Sleep(time.Duration(model.ConsumerRetryDelaySeconds) * time.Second)
		}

		processingError = processEvent(
			requestContext,
			databaseConnection,
			redisClient,
			eventEnvelope.Event,
			deliveryMessage.Body,
		)
		if processingError == nil {
			ackMessage(deliveryMessage, eventEnvelope.Event, eventEnvelope.RequestID)
			return
		}

		if errors.Is(processingError, model.ErrInvalidEventPayload) {
			log.Printf("[consumer] validasi payload gagal. event=%s request_id=%s error=%v", eventEnvelope.Event, eventEnvelope.RequestID, processingError)
			nackMessage(deliveryMessage, "payload invalid", false)
			return
		}
	}

	log.Printf("[consumer] proses gagal setelah retry maksimal. event=%s request_id=%s error=%v", eventEnvelope.Event, eventEnvelope.RequestID, processingError)
	nackMessage(deliveryMessage, "max retry reached", false)
}

func processEvent(
	requestContext context.Context,
	databaseConnection *sql.DB,
	redisClient *redis.Client,
	eventName string,
	messageBody []byte,
) error {
	switch eventName {
	case model.EventThreadCreated:
		threadCreatedEvent, parseError := module.ParseThreadCreatedEvent(messageBody)
		if parseError != nil {
			return parseError
		}
		return module.ProcessThreadCreatedEvent(requestContext, databaseConnection, redisClient, threadCreatedEvent)
	case model.EventCommentCreated:
		commentCreatedEvent, parseError := module.ParseCommentCreatedEvent(messageBody)
		if parseError != nil {
			return parseError
		}
		return module.ProcessCommentCreatedEvent(requestContext, databaseConnection, redisClient, commentCreatedEvent)
	case model.EventThreadLiked:
		threadLikedEvent, parseError := module.ParseThreadLikedEvent(messageBody)
		if parseError != nil {
			return parseError
		}
		return module.ProcessThreadLikedEvent(requestContext, databaseConnection, redisClient, threadLikedEvent)
	default:
		return model.ErrInvalidEventPayload
	}
}

func ackMessage(
	deliveryMessage amqp.Delivery,
	eventName string,
	requestID string,
) {
	if ackError := deliveryMessage.Ack(false); ackError != nil {
		log.Printf("[consumer] ack gagal. event=%s request_id=%s error=%v", eventName, requestID, ackError)
		return
	}

	log.Printf("[consumer] ack sukses. event=%s request_id=%s", eventName, requestID)
}

func nackMessage(
	deliveryMessage amqp.Delivery,
	reason string,
	requeue bool,
) {
	if nackError := deliveryMessage.Nack(false, requeue); nackError != nil {
		log.Printf("[consumer] nack gagal. reason=%s requeue=%t error=%v", reason, requeue, nackError)
		return
	}

	log.Printf("[consumer] nack sukses. reason=%s requeue=%t", reason, requeue)
}
