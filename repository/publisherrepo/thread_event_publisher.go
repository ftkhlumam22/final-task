package publisherrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"final-task/helper"
	"final-task/messaging"
	"final-task/model"

	amqp "github.com/rabbitmq/amqp091-go"
)

const defaultRPCWaitTimeout = 10 * time.Second

func NewThreadEventPublisher(rabbitPublisher *model.RabbitPublisher) *ThreadEventPublisher {
	return &ThreadEventPublisher{rabbitPublisher: rabbitPublisher}
}

func NewThreadRPCPublisher(rabbitRPCClient *model.RabbitRPCClient) *ThreadRPCPublisher {
	return &ThreadRPCPublisher{rabbitRPCClient: rabbitRPCClient}
}

func (threadEventPublisher *ThreadEventPublisher) PublishThreadCreatedEvent(
	threadCreatedEvent model.ThreadCreatedEvent,
) error {
	return threadEventPublisher.publishEventByName(
		model.EventThreadCreated,
		threadCreatedEvent.RequestID,
		threadCreatedEvent,
	)
}

func (threadEventPublisher *ThreadEventPublisher) PublishCommentCreatedEvent(
	commentCreatedEvent model.CommentCreatedEvent,
) error {
	return threadEventPublisher.publishEventByName(
		model.EventCommentCreated,
		commentCreatedEvent.RequestID,
		commentCreatedEvent,
	)
}

func (threadEventPublisher *ThreadEventPublisher) PublishThreadLikedEvent(
	threadLikedEvent model.ThreadLikedEvent,
) error {
	return threadEventPublisher.publishEventByName(
		model.EventThreadLiked,
		threadLikedEvent.RequestID,
		threadLikedEvent,
	)
}

func (threadRPCPublisher *ThreadRPCPublisher) RPCThreadGetLikedEvent(
	threadLikedGetEvent model.ThreadGetLikedEvent,
) ([]byte, error) {
	return threadRPCPublisher.publishRPCEventByName(
		model.EventThreadGetLiked,
		threadLikedGetEvent.RequestID,
		threadLikedGetEvent,
	)
}

func (threadRPCPublisher *ThreadRPCPublisher) publishRPCEventByName(
	eventName string,
	requestID string,
	eventPayloadData any,
) ([]byte, error) {
	routingKey, err := messaging.RoutingKeyForEvent(eventName)
	if err != nil {
		return nil, err
	}

	requestContext, cancel := context.WithTimeout(context.Background(), defaultRPCWaitTimeout)
	defer cancel()

	eventPayload, err := json.Marshal(eventPayloadData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", model.ErrPublishEvent, err)
	}

	log.Printf(
		"[PRODUCER] Mulai kirim event %s ke RabbitMQ. request_id=%s exchange=%s routing_key=%s",
		eventName,
		requestID,
		threadRPCPublisher.rabbitRPCClient.ExchangeName,
		routingKey,
	)

	consumerTag := fmt.Sprintf("rpc-%s", requestID)
	msgs, err := threadRPCPublisher.rabbitRPCClient.Channel.Consume(
		threadRPCPublisher.rabbitRPCClient.QueueName,
		consumerTag,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("%w: start consume: %v", model.ErrConsumeEvent, err)
	}
	defer func() {
		_ = threadRPCPublisher.rabbitRPCClient.Channel.Cancel(consumerTag, false)
	}()

	corrId := helper.RandomString(31)

	err = threadRPCPublisher.rabbitRPCClient.Channel.PublishWithContext(
		requestContext,
		threadRPCPublisher.rabbitRPCClient.ExchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrId,
			ReplyTo:       threadRPCPublisher.rabbitRPCClient.QueueName,
			Body:          eventPayload,
		},
	)
	if err != nil {
		log.Printf("[PRODUCER] Gagal kirim event %s ke RabbitMQ. request_id=%s err=%v", eventName, requestID, err)
		return nil, fmt.Errorf("%w: %v", model.ErrPublishEvent, err)
	}

	log.Printf("[PRODUCER] Event %s berhasil dikirim ke RabbitMQ. request_id=%s", eventName, requestID)
	for {
		select {
		case <-requestContext.Done():
			return nil, fmt.Errorf("%w: wait rpc response: %v", model.ErrConsumeEvent, requestContext.Err())
		case delivery, isOpen := <-msgs:
			if !isOpen {
				return nil, fmt.Errorf("%w: rpc response channel closed", model.ErrConsumeEvent)
			}

			if delivery.CorrelationId != "" && delivery.CorrelationId != corrId {
				continue
			}

			log.Printf("[PRODUCER] RPC response %s diterima. request_id=%s", eventName, requestID)
			return delivery.Body, nil
		}
	}
}

func (threadEventPublisher *ThreadEventPublisher) publishEvent(
	routingKey string,
	eventName string,
	requestID string,
	eventPayloadData any,
) error {
	requestContext := context.Background()

	eventPayload, err := json.Marshal(eventPayloadData)
	if err != nil {
		return fmt.Errorf("%w: %v", model.ErrPublishEvent, err)
	}

	log.Printf(
		"[PRODUCER] Mulai kirim event %s ke RabbitMQ. request_id=%s exchange=%s routing_key=%s",
		eventName,
		requestID,
		threadEventPublisher.rabbitPublisher.ExchangeName,
		routingKey,
	)

	err = threadEventPublisher.rabbitPublisher.Channel.PublishWithContext(
		requestContext,
		threadEventPublisher.rabbitPublisher.ExchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         eventPayload,
		},
	)
	if err != nil {
		log.Printf("[PRODUCER] Gagal kirim event %s ke RabbitMQ. request_id=%s err=%v", eventName, requestID, err)
		return fmt.Errorf("%w: %v", model.ErrPublishEvent, err)
	}

	log.Printf("[PRODUCER] Event %s berhasil dikirim ke RabbitMQ. request_id=%s", eventName, requestID)
	return nil
}

func (threadEventPublisher *ThreadEventPublisher) publishEventByName(
	eventName string,
	requestID string,
	eventPayloadData any,
) error {
	routingKey, err := messaging.RoutingKeyForEvent(eventName)
	if err != nil {
		return err
	}

	return threadEventPublisher.publishEvent(routingKey, eventName, requestID, eventPayloadData)
}
