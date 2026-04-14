package module

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"kaktus-consumer/model"
)

const publishResponseTimeout = 5 * time.Second

func NewConsumerService(dependency ConsumerServiceDependency) ConsumerService {
	return &consumerService{
		eventService:        dependency.EventService,
		publisherRepository: dependency.PublisherRepository,
		maxRetry:            dependency.MaxRetry,
		retryDelay:          dependency.RetryDelay,
	}
}

func (service *consumerService) HandleThreadCreatedMessage(
	requestContext context.Context,
	messageBody []byte,
) ConsumeDecision {
	return service.handleDomainMessage(requestContext, model.EventThreadCreated, messageBody)
}

func (service *consumerService) HandleCommentCreatedMessage(
	requestContext context.Context,
	messageBody []byte,
) ConsumeDecision {
	return service.handleDomainMessage(requestContext, model.EventCommentCreated, messageBody)
}

func (service *consumerService) HandleThreadLikedMessage(
	requestContext context.Context,
	messageBody []byte,
) ConsumeDecision {
	return service.handleDomainMessage(requestContext, model.EventThreadLiked, messageBody)
}

func (service *consumerService) HandleThreadGetLikedMessage(
	requestContext context.Context,
	messageBody []byte,
	replyToQueue string,
	correlationID string,
) ConsumeDecision {
	eventEnvelope, err := service.eventService.ParseEventEnvelope(messageBody)
	if err != nil {
		return ConsumeDecision{
			EventName:  model.EventThreadGetLiked,
			ShouldAck:  false,
			Requeue:    false,
			NackReason: "payload invalid",
			Err:        err,
		}
	}

	requestID := strings.TrimSpace(eventEnvelope.RequestID)
	if eventEnvelope.Event != model.EventThreadGetLiked {
		return ConsumeDecision{
			EventName:  model.EventThreadGetLiked,
			RequestID:  requestID,
			ShouldAck:  false,
			Requeue:    false,
			NackReason: "payload invalid",
			Err:        model.ErrInvalidEventPayload,
		}
	}

	responsePayload, processingErr := service.buildThreadGetLikedResponseWithRetry(
		requestContext,
		messageBody,
		requestID,
	)
	if processingErr != nil {
		return ConsumeDecision{
			EventName:  model.EventThreadGetLiked,
			RequestID:  requestID,
			ShouldAck:  false,
			Requeue:    false,
			NackReason: "max retry reached",
			Err:        processingErr,
		}
	}

	publishContext, cancelPublish := context.WithTimeout(requestContext, publishResponseTimeout)
	defer cancelPublish()

	err = service.publisherRepository.PublishRPCResponse(
		publishContext,
		replyToQueue,
		correlationID,
		responsePayload,
	)
	if err != nil {
		return ConsumeDecision{
			EventName:  model.EventThreadGetLiked,
			RequestID:  requestID,
			ShouldAck:  false,
			Requeue:    false,
			NackReason: "publish rpc response failed",
			Err:        err,
		}
	}

	return ConsumeDecision{
		EventName: model.EventThreadGetLiked,
		RequestID: requestID,
		ShouldAck: true,
	}
}

func (service *consumerService) handleDomainMessage(
	requestContext context.Context,
	expectedEventName string,
	messageBody []byte,
) ConsumeDecision {
	eventEnvelope, err := service.eventService.ParseEventEnvelope(messageBody)
	if err != nil {
		return ConsumeDecision{
			EventName:  expectedEventName,
			ShouldAck:  false,
			Requeue:    false,
			NackReason: "payload invalid",
			Err:        err,
		}
	}

	requestID := strings.TrimSpace(eventEnvelope.RequestID)
	if eventEnvelope.Event != expectedEventName {
		return ConsumeDecision{
			EventName:  expectedEventName,
			RequestID:  requestID,
			ShouldAck:  false,
			Requeue:    false,
			NackReason: "payload invalid",
			Err:        model.ErrInvalidEventPayload,
		}
	}

	err = service.processEventWithRetry(
		requestContext,
		expectedEventName,
		requestID,
		messageBody,
	)
	if err != nil {
		nackReason := "max retry reached"
		if errors.Is(err, model.ErrInvalidEventPayload) {
			nackReason = "payload invalid"
		}

		return ConsumeDecision{
			EventName:  expectedEventName,
			RequestID:  requestID,
			ShouldAck:  false,
			Requeue:    false,
			NackReason: nackReason,
			Err:        err,
		}
	}

	return ConsumeDecision{
		EventName: expectedEventName,
		RequestID: requestID,
		ShouldAck: true,
	}
}

func (service *consumerService) processEventWithRetry(
	requestContext context.Context,
	eventName string,
	requestID string,
	messageBody []byte,
) error {
	var err error
	for retryAttempt := 0; retryAttempt <= service.maxRetry; retryAttempt++ {
		if retryAttempt > 0 {
			time.Sleep(service.retryDelay)
		}

		err = service.eventService.ProcessEvent(requestContext, eventName, messageBody)
		if err == nil {
			return nil
		}
		if errors.Is(err, model.ErrInvalidEventPayload) {
			return err
		}
	}

	return fmt.Errorf("%w: event=%s request_id=%s", err, eventName, requestID)
}

func (service *consumerService) buildThreadGetLikedResponseWithRetry(
	requestContext context.Context,
	messageBody []byte,
	requestID string,
) (interface{}, error) {
	var (
		responsePayload interface{}
		err             error
	)

	for retryAttempt := 0; retryAttempt <= service.maxRetry; retryAttempt++ {
		if retryAttempt > 0 {
			time.Sleep(service.retryDelay)
		}

		responsePayload, err = service.eventService.BuildThreadGetLikedResponse(requestContext, messageBody)
		if err == nil {
			return responsePayload, nil
		}

		if errors.Is(err, model.ErrInvalidEventPayload) {
			return nil, err
		}

		if errors.Is(err, model.ErrConsumeEvent) {
			return model.ThreadGetLikedErrorResponse{
				Error: model.MessageFailedFetchLikedThreads,
			}, nil
		}
	}

	return nil, fmt.Errorf("%w: event=%s request_id=%s", err, model.EventThreadGetLiked, requestID)
}
