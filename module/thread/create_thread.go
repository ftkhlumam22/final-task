package threadmodule

import (
	"log"

	"final-task/helper"
	"final-task/model"
)

func (threadService *ThreadService) CreateThread(
	threadTitle string,
	threadDescription string,
	createdBy int64,
) error {
	requestID, err := helper.NewRequestID()
	if err != nil {
		return err
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

	err = threadService.threadEventPublisher.PublishThreadCreatedEvent(threadCreatedEvent)
	if err != nil {
		log.Printf("[PRODUCER] Publish event thread.created gagal. request_id=%s err=%v", requestID, err)
		return err
	}

	log.Printf("[PRODUCER] Publish event thread.created berhasil. request_id=%s", requestID)

	err = threadService.invalidateThreadListCache()
	if err != nil {
		log.Printf("[PRODUCER] Gagal hapus cache list thread. request_id=%s err=%v", requestID, err)
		return err
	}
	log.Printf("[PRODUCER] Cache list thread berhasil dihapus. request_id=%s", requestID)

	return nil
}
