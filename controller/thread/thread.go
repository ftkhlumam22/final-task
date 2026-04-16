package thread

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"kaktus-consumer/model"
	"kaktus-consumer/module"

	amqp "github.com/rabbitmq/amqp091-go"
)

type deliveryHandler func(requestContext context.Context, deliveryMessage amqp.Delivery) module.ConsumeDecision

func NewController(dependency Dependency) Controller {
	return Controller{
		threadService:           dependency.ThreadService,
		subscriberRepository:    dependency.SubscriberRepository,
		threadCreatedQueueName:  dependency.ThreadCreatedQueueName,
		commentCreatedQueueName: dependency.CommentCreatedQueueName,
		threadLikedQueueName:    dependency.ThreadLikedQueueName,
		threadGetLikedQueueName: dependency.ThreadGetLikedQueueName,
	}
}

func (controller Controller) Start(requestContext context.Context) error {
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

func (controller Controller) consumeQueue(
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

func (controller Controller) startThreadCreatedConsumer(
	requestContext context.Context,
	errChannel chan<- error,
) {
	err := controller.consumeQueue(
		requestContext,
		controller.threadCreatedQueueName,
		threadCreatedConsumerTag,
		func(requestContext context.Context, deliveryMessage amqp.Delivery) module.ConsumeDecision {
			threadCreatedEvent, err := parseThreadCreatedEvent(deliveryMessage.Body)
			if err != nil {
				return invalidPayloadDecision(model.EventThreadCreated, "", err)
			}
			return controller.threadService.HandleThreadCreatedMessage(requestContext, threadCreatedEvent)
		},
	)
	if err != nil {
		log.Printf("[consumer] thread.created consumer berhenti. error=%v", err)
		errChannel <- err
	}
}

func (controller Controller) startCommentCreatedConsumer(
	requestContext context.Context,
	errChannel chan<- error,
) {
	err := controller.consumeQueue(
		requestContext,
		controller.commentCreatedQueueName,
		commentCreatedConsumerTag,
		func(requestContext context.Context, deliveryMessage amqp.Delivery) module.ConsumeDecision {
			commentCreatedEvent, err := parseCommentCreatedEvent(deliveryMessage.Body)
			if err != nil {
				return invalidPayloadDecision(model.EventCommentCreated, "", err)
			}
			return controller.threadService.HandleCommentCreatedMessage(requestContext, commentCreatedEvent)
		},
	)
	if err != nil {
		log.Printf("[consumer] comment.created consumer berhenti. error=%v", err)
		errChannel <- err
	}
}

func (controller Controller) startThreadLikedConsumer(
	requestContext context.Context,
	errChannel chan<- error,
) {
	err := controller.consumeQueue(
		requestContext,
		controller.threadLikedQueueName,
		threadLikedConsumerTag,
		func(requestContext context.Context, deliveryMessage amqp.Delivery) module.ConsumeDecision {
			threadLikedEvent, err := parseThreadLikedEvent(deliveryMessage.Body)
			if err != nil {
				return invalidPayloadDecision(model.EventThreadLiked, "", err)
			}
			return controller.threadService.HandleThreadLikedMessage(requestContext, threadLikedEvent)
		},
	)
	if err != nil {
		log.Printf("[consumer] thread.liked consumer berhenti. error=%v", err)
		errChannel <- err
	}
}

func (controller Controller) startThreadGetLikedConsumer(
	requestContext context.Context,
	errChannel chan<- error,
) {
	err := controller.consumeQueue(
		requestContext,
		controller.threadGetLikedQueueName,
		threadGetLikedConsumerTag,
		func(requestContext context.Context, deliveryMessage amqp.Delivery) module.ConsumeDecision {
			threadGetLikedRequest, err := parseThreadGetLikedRequest(deliveryMessage.Body)
			if err != nil {
				return invalidPayloadDecision(model.EventThreadGetLiked, "", err)
			}

			replyToQueue := strings.TrimSpace(deliveryMessage.ReplyTo)
			if replyToQueue == "" {
				return invalidPayloadDecision(model.EventThreadGetLiked, threadGetLikedRequest.RequestID, model.ErrInvalidEventPayload)
			}

			correlationID := strings.TrimSpace(deliveryMessage.CorrelationId)
			return controller.threadService.HandleThreadGetLikedMessage(
				requestContext,
				threadGetLikedRequest,
				replyToQueue,
				correlationID,
			)
		},
	)
	if err != nil {
		log.Printf("[consumer] thread.get.liked consumer berhenti. error=%v", err)
		errChannel <- err
	}
}

func (controller Controller) applyDecision(
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

func (controller Controller) ackMessage(
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

func (controller Controller) nackMessage(
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

func parseThreadCreatedEvent(messageBody []byte) (model.ThreadCreatedEvent, error) {
	var threadCreatedEvent model.ThreadCreatedEvent
	err := json.Unmarshal(messageBody, &threadCreatedEvent)
	if err != nil {
		return model.ThreadCreatedEvent{}, model.ErrInvalidEventPayload
	}

	threadCreatedEvent.Event = strings.TrimSpace(threadCreatedEvent.Event)
	threadCreatedEvent.RequestID = strings.TrimSpace(threadCreatedEvent.RequestID)
	if threadCreatedEvent.Event != model.EventThreadCreated || threadCreatedEvent.RequestID == "" {
		return model.ThreadCreatedEvent{}, model.ErrInvalidEventPayload
	}

	return threadCreatedEvent, nil
}

func parseCommentCreatedEvent(messageBody []byte) (model.CommentCreatedEvent, error) {
	var commentCreatedEvent model.CommentCreatedEvent
	err := json.Unmarshal(messageBody, &commentCreatedEvent)
	if err != nil {
		return model.CommentCreatedEvent{}, model.ErrInvalidEventPayload
	}

	commentCreatedEvent.Event = strings.TrimSpace(commentCreatedEvent.Event)
	commentCreatedEvent.RequestID = strings.TrimSpace(commentCreatedEvent.RequestID)
	if commentCreatedEvent.Event != model.EventCommentCreated || commentCreatedEvent.RequestID == "" {
		return model.CommentCreatedEvent{}, model.ErrInvalidEventPayload
	}

	return commentCreatedEvent, nil
}

func parseThreadLikedEvent(messageBody []byte) (model.ThreadLikedEvent, error) {
	var threadLikedEvent model.ThreadLikedEvent
	err := json.Unmarshal(messageBody, &threadLikedEvent)
	if err != nil {
		return model.ThreadLikedEvent{}, model.ErrInvalidEventPayload
	}

	threadLikedEvent.Event = strings.TrimSpace(threadLikedEvent.Event)
	threadLikedEvent.RequestID = strings.TrimSpace(threadLikedEvent.RequestID)
	if threadLikedEvent.Event != model.EventThreadLiked || threadLikedEvent.RequestID == "" {
		return model.ThreadLikedEvent{}, model.ErrInvalidEventPayload
	}

	return threadLikedEvent, nil
}

func parseThreadGetLikedRequest(messageBody []byte) (model.ThreadGetLikedRequest, error) {
	var threadGetLikedRequest model.ThreadGetLikedRequest
	err := json.Unmarshal(messageBody, &threadGetLikedRequest)
	if err != nil {
		return model.ThreadGetLikedRequest{}, model.ErrInvalidEventPayload
	}

	threadGetLikedRequest.Event = strings.TrimSpace(threadGetLikedRequest.Event)
	threadGetLikedRequest.RequestID = strings.TrimSpace(threadGetLikedRequest.RequestID)
	if threadGetLikedRequest.Event != model.EventThreadGetLiked || threadGetLikedRequest.RequestID == "" {
		return model.ThreadGetLikedRequest{}, model.ErrInvalidEventPayload
	}

	return threadGetLikedRequest, nil
}
