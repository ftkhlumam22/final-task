package controller

import (
	"context"
	"log"

	"kaktus-consumer/module"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (controller *EventConsumerController) startThreadGetLikedConsumer(
	requestContext context.Context,
	errChannel chan<- error,
) {
	err := controller.consumeQueue(
		requestContext,
		controller.threadGetLikedQueueName,
		threadGetLikedConsumerTag,
		func(requestContext context.Context, deliveryMessage amqp.Delivery) module.ConsumeDecision {
			return controller.consumerService.HandleThreadGetLikedMessage(
				requestContext,
				deliveryMessage.Body,
				deliveryMessage.ReplyTo,
				deliveryMessage.CorrelationId,
			)
		},
	)
	if err != nil {
		log.Printf("[consumer] thread.get.liked consumer berhenti. error=%v", err)
		errChannel <- err
	}
}
