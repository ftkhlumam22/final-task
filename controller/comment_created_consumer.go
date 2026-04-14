package controller

import (
	"context"
	"log"

	"kaktus-consumer/module"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (controller *EventConsumerController) startCommentCreatedConsumer(
	requestContext context.Context,
	errChannel chan<- error,
) {
	err := controller.consumeQueue(
		requestContext,
		controller.commentCreatedQueueName,
		commentCreatedConsumerTag,
		func(requestContext context.Context, deliveryMessage amqp.Delivery) module.ConsumeDecision {
			return controller.consumerService.HandleCommentCreatedMessage(requestContext, deliveryMessage.Body)
		},
	)
	if err != nil {
		log.Printf("[consumer] comment.created consumer berhenti. error=%v", err)
		errChannel <- err
	}
}
