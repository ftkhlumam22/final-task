package module

import (
	"context"
	"log"

	"final-task/helper"
	"final-task/model"
	"final-task/repository"
	"github.com/redis/go-redis/v9"
)

func CreateThread(
	requestContext context.Context,
	rabbitPublisher *model.RabbitPublisher,
	redisClient *redis.Client,
	threadTitle string,
	threadDescription string,
	createdBy int64,
) error {
	requestID, requestIDError := helper.NewRequestID()
	if requestIDError != nil {
		return requestIDError
	}
	log.Printf("[PRODUCER] Request ID thread dibuat. request_id=%s", requestID)

	threadCreatedEvent := model.ThreadCreatedEvent{
		Event:       model.EventThreadCreated,
		RequestID:   requestID,
		CreatedBy:   createdBy,
		Title:       threadTitle,
		Description: threadDescription,
	}
	log.Printf("[PRODUCER] Menyiapkan event thread.created. request_id=%s", requestID)

	publishError := repository.PublishThreadCreatedEvent(requestContext, rabbitPublisher, threadCreatedEvent)
	if publishError != nil {
		log.Printf("[PRODUCER] Publish event thread.created gagal. request_id=%s err=%v", requestID, publishError)
		return publishError
	}

	log.Printf("[PRODUCER] Publish event thread.created berhasil. request_id=%s", requestID)

	cacheInvalidateError := repository.InvalidateThreadListCache(requestContext, redisClient)
	if cacheInvalidateError != nil {
		log.Printf("[PRODUCER] Gagal hapus cache list thread. request_id=%s err=%v", requestID, cacheInvalidateError)
		return cacheInvalidateError
	}
	log.Printf("[PRODUCER] Cache list thread berhasil dihapus. request_id=%s", requestID)

	return nil
}
