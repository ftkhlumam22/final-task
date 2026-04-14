package controller

import (
	"context"
	"log"

	"kaktus-consumer/module"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (controller *EventConsumerController) startThreadCreatedConsumer(
	requestContext context.Context,
	errChannel chan<- error,
) {
	err := controller.consumeQueue(
		requestContext,
		controller.threadCreatedQueueName,
		threadCreatedConsumerTag,
		func(requestContext context.Context, deliveryMessage amqp.Delivery) module.ConsumeDecision {
			return controller.consumerService.HandleThreadCreatedMessage(requestContext, deliveryMessage.Body)
		},
	)
	if err != nil {
		log.Printf("[consumer] thread.created consumer berhenti. error=%v", err)
		errChannel <- err
	}
}
