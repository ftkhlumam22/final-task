package controller

import (
	"context"
	"log"

	"kaktus-consumer/module"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (controller *EventConsumerController) startThreadLikedConsumer(
	requestContext context.Context,
	errChannel chan<- error,
) {
	err := controller.consumeQueue(
		requestContext,
		controller.threadLikedQueueName,
		threadLikedConsumerTag,
		func(requestContext context.Context, deliveryMessage amqp.Delivery) module.ConsumeDecision {
			return controller.consumerService.HandleThreadLikedMessage(requestContext, deliveryMessage.Body)
		},
	)
	if err != nil {
		log.Printf("[consumer] thread.liked consumer berhenti. error=%v", err)
		errChannel <- err
	}
}
