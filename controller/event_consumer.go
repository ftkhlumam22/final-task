package controller

import (
	"context"
	"fmt"
	"log"

	"kaktus-consumer/model"
	"kaktus-consumer/module"

	amqp "github.com/rabbitmq/amqp091-go"
)

type deliveryHandler func(requestContext context.Context, deliveryMessage amqp.Delivery) module.ConsumeDecision

func NewEventConsumerController(dependency EventConsumerControllerDependency) *EventConsumerController {
	return &EventConsumerController{
		consumerService:        dependency.ConsumerService,
		subscriberRepository:   dependency.SubscriberRepository,
		threadCreatedQueueName: dependency.ThreadCreatedQueueName,
		commentCreatedQueueName: dependency.CommentCreatedQueueName,
		threadLikedQueueName:   dependency.ThreadLikedQueueName,
		threadGetLikedQueueName: dependency.ThreadGetLikedQueueName,
	}
}

func (controller *EventConsumerController) Start(requestContext context.Context) error {
	errChannel := make(chan error, 4)

	go controller.startThreadCreatedConsumer(requestContext, errChannel)
	go controller.startCommentCreatedConsumer(requestContext, errChannel)
	go controller.startThreadLikedConsumer(requestContext, errChannel)
	go controller.startThreadGetLikedConsumer(requestContext, errChannel)

	log.Printf(
		"[consumer] worker aktif. thread_created_queue=%s comment_created_queue=%s thread_liked_queue=%s thread_get_liked_queue=%s",
		controller.threadCreatedQueueName,
		controller.commentCreatedQueueName,
		controller.threadLikedQueueName,
		controller.threadGetLikedQueueName,
	)

	select {
	case <-requestContext.Done():
		log.Printf("[consumer] worker berhenti: context selesai")
		return nil
	case err := <-errChannel:
		return err
	}
}

func (controller *EventConsumerController) consumeQueue(
	requestContext context.Context,
	queueName string,
	consumerTag string,
	handler deliveryHandler,
) error {
	deliveryChannel, err := controller.subscriberRepository.Consume(queueName, consumerTag)
	if err != nil {
		return err
	}

	for {
		select {
		case <-requestContext.Done():
			return nil
		case deliveryMessage, isChannelOpen := <-deliveryChannel:
			if !isChannelOpen {
				return fmt.Errorf("%w: rabbit delivery channel closed. queue=%s", model.ErrConsumeEvent, queueName)
			}

			decision := handler(requestContext, deliveryMessage)
			controller.applyDecision(deliveryMessage, decision)
		}
	}
}

func (controller *EventConsumerController) applyDecision(
	deliveryMessage amqp.Delivery,
	decision module.ConsumeDecision,
) {
	if decision.ShouldAck {
		controller.ackMessage(deliveryMessage, decision.EventName, decision.RequestID)
		return
	}

	if decision.Err != nil {
		log.Printf(
			"[consumer] proses gagal. event=%s request_id=%s error=%v",
			decision.EventName,
			decision.RequestID,
			decision.Err,
		)
	}

	controller.nackMessage(deliveryMessage, decision.NackReason, decision.Requeue)
}

func (controller *EventConsumerController) ackMessage(
	deliveryMessage amqp.Delivery,
	eventName string,
	requestID string,
) {
	err := controller.subscriberRepository.Ack(deliveryMessage)
	if err != nil {
		log.Printf("[consumer] ack gagal. event=%s request_id=%s error=%v", eventName, requestID, err)
		return
	}

	log.Printf("[consumer] ack sukses. event=%s request_id=%s", eventName, requestID)
}

func (controller *EventConsumerController) nackMessage(
	deliveryMessage amqp.Delivery,
	reason string,
	requeue bool,
) {
	err := controller.subscriberRepository.Nack(deliveryMessage, requeue)
	if err != nil {
		log.Printf("[consumer] nack gagal. reason=%s requeue=%t error=%v", reason, requeue, err)
		return
	}

	log.Printf("[consumer] nack sukses. reason=%s requeue=%t", reason, requeue)
}
