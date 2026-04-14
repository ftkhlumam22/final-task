package module

import (
	"context"
	"fmt"
	"strings"

	"kaktus-consumer/helper"
	"kaktus-consumer/model"
	"kaktus-consumer/repository"
)

const cacheScanBatchSize = 100

func NewEventService(dependency EventServiceDependency) EventService {
	return &eventService{
		sqlRepository:   dependency.SQLRepository,
		cacheRepository: dependency.CacheRepository,
	}
}

func (service *eventService) ProcessEvent(
	requestContext context.Context,
	eventName string,
	messageBody []byte,
) error {
	switch eventName {
	case model.EventThreadCreated:
		threadCreatedEvent, err := parseThreadCreatedEvent(messageBody)
		if err != nil {
			return err
		}
		return service.processThreadCreatedEvent(requestContext, threadCreatedEvent)
	case model.EventCommentCreated:
		commentCreatedEvent, err := parseCommentCreatedEvent(messageBody)
		if err != nil {
			return err
		}
		return service.processCommentCreatedEvent(requestContext, commentCreatedEvent)
	case model.EventThreadLiked:
		threadLikedEvent, err := parseThreadLikedEvent(messageBody)
		if err != nil {
			return err
		}
		return service.processThreadLikedEvent(requestContext, threadLikedEvent)
	default:
		return model.ErrInvalidEventPayload
	}
}

func (service *eventService) BuildThreadGetLikedResponse(
	requestContext context.Context,
	messageBody []byte,
) (interface{}, error) {
	threadGetLikedRequest, err := parseThreadGetLikedRequest(messageBody)
	if err != nil {
		return nil, err
	}

	threadGetLikedRequest.Event = strings.TrimSpace(threadGetLikedRequest.Event)
	threadGetLikedRequest.RequestID = strings.TrimSpace(threadGetLikedRequest.RequestID)
	if err = validateThreadGetLikedRequest(threadGetLikedRequest); err != nil {
		return nil, err
	}

	likedThreads, err := service.sqlRepository.GetLikedThreads(requestContext)
	if err != nil {
		return nil, err
	}

	return model.ThreadGetLikedSuccessResponse{Data: likedThreads}, nil
}

func (service *eventService) processThreadCreatedEvent(
	requestContext context.Context,
	threadCreatedEvent model.ThreadCreatedEvent,
) error {
	threadCreatedEvent.Event = strings.TrimSpace(threadCreatedEvent.Event)
	threadCreatedEvent.RequestID = strings.TrimSpace(threadCreatedEvent.RequestID)
	threadCreatedEvent.Title = strings.TrimSpace(threadCreatedEvent.Title)
	threadCreatedEvent.Description = strings.TrimSpace(threadCreatedEvent.Description)

	if err := validateThreadCreatedEvent(threadCreatedEvent); err != nil {
		return err
	}

	_, err := service.sqlRepository.InsertThread(requestContext, threadCreatedEvent)
	if err != nil {
		return err
	}

	return service.invalidateThreadListCache(requestContext)
}

func (service *eventService) processCommentCreatedEvent(
	requestContext context.Context,
	commentCreatedEvent model.CommentCreatedEvent,
) error {
	commentCreatedEvent.Event = strings.TrimSpace(commentCreatedEvent.Event)
	commentCreatedEvent.RequestID = strings.TrimSpace(commentCreatedEvent.RequestID)
	commentCreatedEvent.Comment = strings.TrimSpace(commentCreatedEvent.Comment)

	if err := validateCommentCreatedEvent(commentCreatedEvent); err != nil {
		return err
	}

	err := service.sqlRepository.WithTransaction(
		requestContext,
		func(sqlTransactionRepository repository.SQLTransactionRepository) error {
			err := sqlTransactionRepository.InsertComment(requestContext, commentCreatedEvent)
			if err != nil {
				return err
			}

			if commentCreatedEvent.ParentCommentID != nil {
				rowsAffected, err := sqlTransactionRepository.IncrementParentCommentReply(
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
				return nil
			}

			rowsAffected, err := sqlTransactionRepository.IncrementThreadTotalComment(
				requestContext,
				commentCreatedEvent.ThreadID,
			)
			if err != nil {
				return err
			}
			if rowsAffected == 0 {
				return fmt.Errorf("%w: update thread total_comment: thread not found", model.ErrConsumeEvent)
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	return service.invalidateThreadDetailCache(requestContext, commentCreatedEvent.ThreadID)
}

func (service *eventService) processThreadLikedEvent(
	requestContext context.Context,
	threadLikedEvent model.ThreadLikedEvent,
) error {
	threadLikedEvent.Event = strings.TrimSpace(threadLikedEvent.Event)
	threadLikedEvent.RequestID = strings.TrimSpace(threadLikedEvent.RequestID)

	if err := validateThreadLikedEvent(threadLikedEvent); err != nil {
		return err
	}

	err := service.sqlRepository.WithTransaction(
		requestContext,
		func(sqlTransactionRepository repository.SQLTransactionRepository) error {
			insertedRows, err := sqlTransactionRepository.InsertThreadLike(requestContext, threadLikedEvent)
			if err != nil {
				return err
			}

			if insertedRows == 0 {
				return nil
			}

			rowsAffected, err := sqlTransactionRepository.IncrementThreadTotalLike(
				requestContext,
				threadLikedEvent.ThreadID,
			)
			if err != nil {
				return err
			}
			if rowsAffected == 0 {
				return fmt.Errorf("%w: update thread total_likes: thread not found", model.ErrConsumeEvent)
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	return service.invalidateThreadDetailCache(requestContext, threadLikedEvent.ThreadID)
}

func (service *eventService) invalidateThreadListCache(requestContext context.Context) error {
	cacheKeys, err := service.cacheRepository.ScanKeys(requestContext, model.ThreadListCachePattern, cacheScanBatchSize)
	if err != nil {
		return err
	}

	if len(cacheKeys) == 0 {
		return nil
	}

	return service.cacheRepository.Delete(requestContext, cacheKeys...)
}

func (service *eventService) invalidateThreadDetailCache(requestContext context.Context, threadID int64) error {
	cacheKey := helper.BuildThreadDetailCacheKey(threadID)
	return service.cacheRepository.Delete(requestContext, cacheKey)
}
