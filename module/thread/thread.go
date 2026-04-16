package thread

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"kaktus-consumer/model"
)

const publishResponseTimeout = 5 * time.Second

func NewService(dependency Dependency) Service {
	return service{
		threadRepository:    dependency.ThreadRepository,
		publisherRepository: dependency.PublisherRepository,
		maxRetry:            dependency.MaxRetry,
		retryDelay:          dependency.RetryDelay,
	}
}

func (service service) HandleThreadCreatedMessage(
	requestContext context.Context,
	threadCreatedEvent model.ThreadCreatedEvent,
) ConsumeDecision {
	threadCreatedEvent = normalizeThreadCreatedEvent(threadCreatedEvent)
	if err := validateThreadCreatedEvent(threadCreatedEvent); err != nil {
		return ConsumeDecision{
			EventName:  model.EventThreadCreated,
			RequestID:  threadCreatedEvent.RequestID,
			ShouldAck:  false,
			Requeue:    false,
			NackReason: "payload invalid",
			Err:        err,
		}
	}

	err := service.processEventWithRetry(
		requestContext,
		model.EventThreadCreated,
		threadCreatedEvent.RequestID,
		func() error {
			return service.processThreadCreatedEvent(requestContext, threadCreatedEvent)
		},
	)
	if err != nil {
		return ConsumeDecision{
			EventName:  model.EventThreadCreated,
			RequestID:  threadCreatedEvent.RequestID,
			ShouldAck:  false,
			Requeue:    false,
			NackReason: "max retry reached",
			Err:        err,
		}
	}

	return ConsumeDecision{
		EventName: model.EventThreadCreated,
		RequestID: threadCreatedEvent.RequestID,
		ShouldAck: true,
	}
}

func (service service) HandleCommentCreatedMessage(
	requestContext context.Context,
	commentCreatedEvent model.CommentCreatedEvent,
) ConsumeDecision {
	commentCreatedEvent = normalizeCommentCreatedEvent(commentCreatedEvent)
	if err := validateCommentCreatedEvent(commentCreatedEvent); err != nil {
		return ConsumeDecision{
			EventName:  model.EventCommentCreated,
			RequestID:  commentCreatedEvent.RequestID,
			ShouldAck:  false,
			Requeue:    false,
			NackReason: "payload invalid",
			Err:        err,
		}
	}

	err := service.processEventWithRetry(
		requestContext,
		model.EventCommentCreated,
		commentCreatedEvent.RequestID,
		func() error {
			return service.processCommentCreatedEvent(requestContext, commentCreatedEvent)
		},
	)
	if err != nil {
		return ConsumeDecision{
			EventName:  model.EventCommentCreated,
			RequestID:  commentCreatedEvent.RequestID,
			ShouldAck:  false,
			Requeue:    false,
			NackReason: "max retry reached",
			Err:        err,
		}
	}

	return ConsumeDecision{
		EventName: model.EventCommentCreated,
		RequestID: commentCreatedEvent.RequestID,
		ShouldAck: true,
	}
}

func (service service) HandleThreadLikedMessage(
	requestContext context.Context,
	threadLikedEvent model.ThreadLikedEvent,
) ConsumeDecision {
	threadLikedEvent = normalizeThreadLikedEvent(threadLikedEvent)
	if err := validateThreadLikedEvent(threadLikedEvent); err != nil {
		return ConsumeDecision{
			EventName:  model.EventThreadLiked,
			RequestID:  threadLikedEvent.RequestID,
			ShouldAck:  false,
			Requeue:    false,
			NackReason: "payload invalid",
			Err:        err,
		}
	}

	err := service.processEventWithRetry(
		requestContext,
		model.EventThreadLiked,
		threadLikedEvent.RequestID,
		func() error {
			return service.processThreadLikedEvent(requestContext, threadLikedEvent)
		},
	)
	if err != nil {
		return ConsumeDecision{
			EventName:  model.EventThreadLiked,
			RequestID:  threadLikedEvent.RequestID,
			ShouldAck:  false,
			Requeue:    false,
			NackReason: "max retry reached",
			Err:        err,
		}
	}

	return ConsumeDecision{
		EventName: model.EventThreadLiked,
		RequestID: threadLikedEvent.RequestID,
		ShouldAck: true,
	}
}

func (service service) HandleThreadGetLikedMessage(
	requestContext context.Context,
	threadGetLikedRequest model.ThreadGetLikedRequest,
	replyToQueue string,
	correlationID string,
) ConsumeDecision {
	threadGetLikedRequest = normalizeThreadGetLikedRequest(threadGetLikedRequest)
	if err := validateThreadGetLikedRequest(threadGetLikedRequest); err != nil {
		return ConsumeDecision{
			EventName:  model.EventThreadGetLiked,
			RequestID:  threadGetLikedRequest.RequestID,
			ShouldAck:  false,
			Requeue:    false,
			NackReason: "payload invalid",
			Err:        err,
		}
	}

	responsePayload, processingErr := service.buildThreadGetLikedResponseWithRetry(
		requestContext,
		threadGetLikedRequest,
	)
	if processingErr != nil {
		return ConsumeDecision{
			EventName:  model.EventThreadGetLiked,
			RequestID:  threadGetLikedRequest.RequestID,
			ShouldAck:  false,
			Requeue:    false,
			NackReason: "max retry reached",
			Err:        processingErr,
		}
	}

	publishContext, cancelPublish := context.WithTimeout(requestContext, publishResponseTimeout)
	defer cancelPublish()

	err := service.publisherRepository.PublishRPCResponse(
		publishContext,
		replyToQueue,
		correlationID,
		responsePayload,
	)
	if err != nil {
		return ConsumeDecision{
			EventName:  model.EventThreadGetLiked,
			RequestID:  threadGetLikedRequest.RequestID,
			ShouldAck:  false,
			Requeue:    false,
			NackReason: "publish rpc response failed",
			Err:        err,
		}
	}

	return ConsumeDecision{
		EventName: model.EventThreadGetLiked,
		RequestID: threadGetLikedRequest.RequestID,
		ShouldAck: true,
	}
}

func (service service) processEventWithRetry(
	requestContext context.Context,
	eventName string,
	requestID string,
	operation func() error,
) error {
	var err error
	for retryAttempt := 0; retryAttempt <= service.maxRetry; retryAttempt++ {
		if retryAttempt > 0 {
			time.Sleep(service.retryDelay)
		}

		err = operation()
		if err == nil {
			return nil
		}
		if errors.Is(err, model.ErrInvalidEventPayload) {
			return err
		}
	}

	return fmt.Errorf("%w: event=%s request_id=%s", err, eventName, requestID)
}

func (service service) buildThreadGetLikedResponseWithRetry(
	requestContext context.Context,
	threadGetLikedRequest model.ThreadGetLikedRequest,
) (interface{}, error) {
	var (
		responsePayload interface{}
		err             error
	)

	for retryAttempt := 0; retryAttempt <= service.maxRetry; retryAttempt++ {
		if retryAttempt > 0 {
			time.Sleep(service.retryDelay)
		}

		responsePayload, err = service.buildThreadGetLikedResponse(requestContext)
		if err == nil {
			return responsePayload, nil
		}

		if errors.Is(err, model.ErrConsumeEvent) {
			return model.ThreadGetLikedErrorResponse{
				Error: model.MessageFailedFetchLikedThreads,
			}, nil
		}
	}

	return nil, fmt.Errorf(
		"%w: event=%s request_id=%s",
		err,
		model.EventThreadGetLiked,
		threadGetLikedRequest.RequestID,
	)
}

func (service service) buildThreadGetLikedResponse(
	requestContext context.Context,
) (interface{}, error) {
	likedThreads, err := service.threadRepository.GetLikedThreads(requestContext)
	if err != nil {
		return nil, err
	}

	return model.ThreadGetLikedSuccessResponse{Data: likedThreads}, nil
}

func (service service) processThreadCreatedEvent(
	requestContext context.Context,
	threadCreatedEvent model.ThreadCreatedEvent,
) error {
	_, err := service.threadRepository.InsertThread(requestContext, threadCreatedEvent)
	if err != nil {
		return err
	}

	return service.threadRepository.InvalidateThreadListCache(requestContext)
}

func (service service) processCommentCreatedEvent(
	requestContext context.Context,
	commentCreatedEvent model.CommentCreatedEvent,
) error {
	transactionRepository, err := service.threadRepository.BeginTransaction(requestContext)
	if err != nil {
		return err
	}
	defer transactionRepository.RollbackTransaction()

	err = transactionRepository.InsertComment(requestContext, commentCreatedEvent)
	if err != nil {
		return err
	}

	if commentCreatedEvent.ParentCommentID != nil {
		rowsAffected, err := transactionRepository.IncrementParentCommentReply(
			requestContext,
			*commentCreatedEvent.ParentCommentID,
			commentCreatedEvent.ThreadID,
		)
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return fmt.Errorf("%w: update parent comment total_reply: parent comment not found", model.ErrInvalidEventPayload)
		}
	} else {
		rowsAffected, err := transactionRepository.IncrementThreadTotalComment(
			requestContext,
			commentCreatedEvent.ThreadID,
		)
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return fmt.Errorf("%w: update thread total_comment: thread not found", model.ErrConsumeEvent)
		}
	}

	err = transactionRepository.CommitTransaction()
	if err != nil {
		return err
	}

	return service.threadRepository.InvalidateThreadDetailCache(requestContext, commentCreatedEvent.ThreadID)
}

func (service service) processThreadLikedEvent(
	requestContext context.Context,
	threadLikedEvent model.ThreadLikedEvent,
) error {
	transactionRepository, err := service.threadRepository.BeginTransaction(requestContext)
	if err != nil {
		return err
	}
	defer transactionRepository.RollbackTransaction()

	insertedRows, err := transactionRepository.InsertThreadLike(requestContext, threadLikedEvent)
	if err != nil {
		return err
	}

	if insertedRows > 0 {
		rowsAffected, err := transactionRepository.IncrementThreadTotalLike(
			requestContext,
			threadLikedEvent.ThreadID,
		)
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return fmt.Errorf("%w: update thread total_likes: thread not found", model.ErrConsumeEvent)
		}
	}

	err = transactionRepository.CommitTransaction()
	if err != nil {
		return err
	}

	return service.threadRepository.InvalidateThreadDetailCache(requestContext, threadLikedEvent.ThreadID)
}

func normalizeThreadCreatedEvent(threadCreatedEvent model.ThreadCreatedEvent) model.ThreadCreatedEvent {
	threadCreatedEvent.Event = strings.TrimSpace(threadCreatedEvent.Event)
	threadCreatedEvent.RequestID = strings.TrimSpace(threadCreatedEvent.RequestID)
	threadCreatedEvent.Title = strings.TrimSpace(threadCreatedEvent.Title)
	threadCreatedEvent.Description = strings.TrimSpace(threadCreatedEvent.Description)
	return threadCreatedEvent
}

func normalizeCommentCreatedEvent(commentCreatedEvent model.CommentCreatedEvent) model.CommentCreatedEvent {
	commentCreatedEvent.Event = strings.TrimSpace(commentCreatedEvent.Event)
	commentCreatedEvent.RequestID = strings.TrimSpace(commentCreatedEvent.RequestID)
	commentCreatedEvent.Comment = strings.TrimSpace(commentCreatedEvent.Comment)
	return commentCreatedEvent
}

func normalizeThreadLikedEvent(threadLikedEvent model.ThreadLikedEvent) model.ThreadLikedEvent {
	threadLikedEvent.Event = strings.TrimSpace(threadLikedEvent.Event)
	threadLikedEvent.RequestID = strings.TrimSpace(threadLikedEvent.RequestID)
	return threadLikedEvent
}

func normalizeThreadGetLikedRequest(threadGetLikedRequest model.ThreadGetLikedRequest) model.ThreadGetLikedRequest {
	threadGetLikedRequest.Event = strings.TrimSpace(threadGetLikedRequest.Event)
	threadGetLikedRequest.RequestID = strings.TrimSpace(threadGetLikedRequest.RequestID)
	return threadGetLikedRequest
}

func validateThreadCreatedEvent(threadCreatedEvent model.ThreadCreatedEvent) error {
	if !isThreadCreatedEventValid(threadCreatedEvent) {
		return model.ErrInvalidEventPayload
	}

	return nil
}

func isThreadCreatedEventValid(threadCreatedEvent model.ThreadCreatedEvent) bool {
	if threadCreatedEvent.RequestID == "" {
		return false
	}
	if threadCreatedEvent.Event != model.EventThreadCreated {
		return false
	}
	if threadCreatedEvent.CreatedBy <= 0 {
		return false
	}

	if strings.TrimSpace(threadCreatedEvent.Title) == "" {
		return false
	}
	if utf8.RuneCountInString(threadCreatedEvent.Title) > model.MaxThreadTitleLength {
		return false
	}

	if strings.TrimSpace(threadCreatedEvent.Description) == "" {
		return false
	}
	if utf8.RuneCountInString(threadCreatedEvent.Description) > model.MaxThreadDescriptionLength {
		return false
	}

	return true
}

func validateCommentCreatedEvent(commentCreatedEvent model.CommentCreatedEvent) error {
	if commentCreatedEvent.Event != model.EventCommentCreated ||
		commentCreatedEvent.RequestID == "" ||
		commentCreatedEvent.ThreadID <= 0 ||
		commentCreatedEvent.CreatedBy <= 0 ||
		commentCreatedEvent.Comment == "" ||
		utf8.RuneCountInString(commentCreatedEvent.Comment) > model.MaxCommentLength {
		return model.ErrInvalidEventPayload
	}

	if commentCreatedEvent.ParentCommentID != nil && *commentCreatedEvent.ParentCommentID <= 0 {
		return model.ErrInvalidEventPayload
	}

	return nil
}

func validateThreadLikedEvent(threadLikedEvent model.ThreadLikedEvent) error {
	if threadLikedEvent.Event != model.EventThreadLiked ||
		threadLikedEvent.RequestID == "" ||
		threadLikedEvent.ThreadID <= 0 ||
		threadLikedEvent.LikedBy <= 0 {
		return model.ErrInvalidEventPayload
	}

	return nil
}

func validateThreadGetLikedRequest(threadGetLikedRequest model.ThreadGetLikedRequest) error {
	if threadGetLikedRequest.Event != model.EventThreadGetLiked ||
		threadGetLikedRequest.RequestID == "" ||
		threadGetLikedRequest.Body == nil {
		return model.ErrInvalidEventPayload
	}

	return nil
}
