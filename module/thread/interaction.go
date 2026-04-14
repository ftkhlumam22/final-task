package threadmodule

import (
	"encoding/json"
	"fmt"
	"log"

	"final-task/dto/response"
	"final-task/helper"
	"final-task/model"
)

func (threadService *ThreadService) CreateComment(
	threadID int64,
	commentText string,
	parentCommentID *int64,
	createdBy int64,
) error {
	requestID, err := helper.NewRequestID()
	if err != nil {
		return err
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

	err = threadService.threadEventPublisher.PublishCommentCreatedEvent(commentCreatedEvent)
	if err != nil {
		log.Printf("[PRODUCER] Publish event comment.created gagal. request_id=%s err=%v", requestID, err)
		return err
	}

	err = threadService.invalidateThreadDetailCache(threadID)
	if err != nil {
		log.Printf("[PRODUCER] Gagal hapus cache detail thread. request_id=%s thread_id=%d err=%v", requestID, threadID, err)
		return err
	}
	log.Printf("[PRODUCER] Cache detail thread berhasil dihapus. request_id=%s thread_id=%d", requestID, threadID)

	return nil
}

func (threadService *ThreadService) GetThreadLiked() ([]response.LikedThreadData, error) {
	requestID, err := helper.NewRequestID()
	if err != nil {
		return nil, err
	}

	getThreadLikedEvent := model.ThreadGetLikedEvent{
		Event:     model.EventThreadGetLiked,
		RequestID: requestID,
		Body:      map[string]any{},
	}
	log.Printf("[PRODUCER] Menyiapkan event thread.get.liked. request_id=%s", requestID)

	messageData, err := threadService.threadRPCPublisher.RPCThreadGetLikedEvent(getThreadLikedEvent)
	if err != nil {
		log.Printf("[PRODUCER] Publish event thread.get.liked gagal. request_id=%s err=%v", requestID, err)
		return nil, err
	}

	var likedThreadList []response.LikedThreadData
	if err := json.Unmarshal(messageData, &likedThreadList); err == nil {
		return likedThreadList, nil
	}

	var wrappedResponse struct {
		Data  []response.LikedThreadData `json:"data"`
		Error string                     `json:"error,omitempty"`
	}
	if err := json.Unmarshal(messageData, &wrappedResponse); err == nil {
		if wrappedResponse.Error != "" {
			return nil, fmt.Errorf("%w: %s", model.ErrConsumeEvent, wrappedResponse.Error)
		}
		if wrappedResponse.Data != nil {
			return wrappedResponse.Data, nil
		}
	}

	return nil, fmt.Errorf("%w: invalid rpc response body", model.ErrConsumeEvent)
}

func (threadService *ThreadService) InsertLikeThread(threadID int64, likedBy int64) error {
	requestID, err := helper.NewRequestID()
	if err != nil {
		return err
	}

	threadLikedEvent := model.ThreadLikedEvent{
		Event:     model.EventThreadLiked,
		RequestID: requestID,
		ThreadID:  threadID,
		LikedBy:   likedBy,
	}
	log.Printf("[PRODUCER] Menyiapkan event thread.liked. request_id=%s thread_id=%d", requestID, threadID)

	err = threadService.threadEventPublisher.PublishThreadLikedEvent(threadLikedEvent)
	if err != nil {
		log.Printf("[PRODUCER] Publish event thread.liked gagal. request_id=%s err=%v", requestID, err)
		return err
	}

	err = threadService.invalidateThreadDetailCache(threadID)
	if err != nil {
		log.Printf("[PRODUCER] Gagal hapus cache detail thread. request_id=%s thread_id=%d err=%v", requestID, threadID, err)
		return err
	}
	log.Printf("[PRODUCER] Cache detail thread berhasil dihapus. request_id=%s thread_id=%d", requestID, threadID)

	return nil
}
