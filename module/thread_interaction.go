package module

import (
	"context"
	"log"

	"final-task/helper"
	"final-task/model"
	"final-task/repository"
	"github.com/redis/go-redis/v9"
)

func CreateComment(
	requestContext context.Context,
	rabbitPublisher *model.RabbitPublisher,
	redisClient *redis.Client,
	threadID int64,
	commentText string,
	parentCommentID *int64,
	createdBy int64,
) error {
	requestID, requestIDError := helper.NewRequestID()
	if requestIDError != nil {
		return requestIDError
	}

	commentCreatedEvent := model.CommentCreatedEvent{
		Event:           model.EventCommentCreated,
		RequestID:       requestID,
		ThreadID:        threadID,
		CreatedBy:       createdBy,
		Comment:         commentText,
		ParentCommentID: parentCommentID,
	}
	log.Printf("[PRODUCER] Menyiapkan event comment.created. request_id=%s thread_id=%d", requestID, threadID)

	publishError := repository.PublishCommentCreatedEvent(requestContext, rabbitPublisher, commentCreatedEvent)
	if publishError != nil {
		log.Printf("[PRODUCER] Publish event comment.created gagal. request_id=%s err=%v", requestID, publishError)
		return publishError
	}

	cacheInvalidateError := repository.InvalidateThreadDetailCache(requestContext, redisClient, threadID)
	if cacheInvalidateError != nil {
		log.Printf("[PRODUCER] Gagal hapus cache detail thread. request_id=%s thread_id=%d err=%v", requestID, threadID, cacheInvalidateError)
		return cacheInvalidateError
	}
	log.Printf("[PRODUCER] Cache detail thread berhasil dihapus. request_id=%s thread_id=%d", requestID, threadID)

	return nil
}

func InsertLikeThread(
	requestContext context.Context,
	rabbitPublisher *model.RabbitPublisher,
	redisClient *redis.Client,
	threadID int64,
	likedBy int64,
) error {
	requestID, requestIDError := helper.NewRequestID()
	if requestIDError != nil {
		return requestIDError
	}

	threadLikedEvent := model.ThreadLikedEvent{
		Event:     model.EventThreadLiked,
		RequestID: requestID,
		ThreadID:  threadID,
		LikedBy:   likedBy,
	}
	log.Printf("[PRODUCER] Menyiapkan event thread.liked. request_id=%s thread_id=%d", requestID, threadID)

	publishError := repository.PublishThreadLikedEvent(requestContext, rabbitPublisher, threadLikedEvent)
	if publishError != nil {
		log.Printf("[PRODUCER] Publish event thread.liked gagal. request_id=%s err=%v", requestID, publishError)
		return publishError
	}

	cacheInvalidateError := repository.InvalidateThreadDetailCache(requestContext, redisClient, threadID)
	if cacheInvalidateError != nil {
		log.Printf("[PRODUCER] Gagal hapus cache detail thread. request_id=%s thread_id=%d err=%v", requestID, threadID, cacheInvalidateError)
		return cacheInvalidateError
	}
	log.Printf("[PRODUCER] Cache detail thread berhasil dihapus. request_id=%s thread_id=%d", requestID, threadID)

	return nil
}
