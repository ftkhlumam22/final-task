package threadmodule

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"final-task/dto/response"
	"final-task/helper"
	"final-task/messaging"
	"final-task/model"

	amqp "github.com/rabbitmq/amqp091-go"
)

const defaultRPCWaitTimeout = 10 * time.Second

func NewThreadService(
	threadSQLRepository ThreadSQLRepository,
	threadListCacheRepository ThreadListCacheRepository,
	threadDetailCacheRepository ThreadDetailCacheRepository,
	rabbitPublisher model.RabbitPublisher,
	rabbitRPCClient model.RabbitRPCClient,
) ThreadService {
	return ThreadService{
		threadSQLRepository:         threadSQLRepository,
		threadListCacheRepository:   threadListCacheRepository,
		threadDetailCacheRepository: threadDetailCacheRepository,
		rabbitPublisher:             rabbitPublisher,
		rabbitRPCClient:             rabbitRPCClient,
	}
}

func (threadService ThreadService) ListThread(page int, limit int) (response.GetAllThread, error) {
	cacheKey := buildThreadListCacheKey(limit)
	cacheField := buildThreadListCacheField(page)

	cachedThreadListValue, isCacheHit, err := threadService.threadListCacheRepository.GetHashField(cacheKey, cacheField)
	if err != nil {
		log.Printf("[PRODUCER] Gagal baca cache list thread. page=%d limit=%d err=%v", page, limit, err)
	}
	if isCacheHit {
		var cachedThreadList response.GetAllThread
		if err = json.Unmarshal([]byte(cachedThreadListValue), &cachedThreadList); err == nil {
			log.Printf("[PRODUCER] List thread diambil dari cache. page=%d limit=%d", page, limit)
			return cachedThreadList, nil
		}
		log.Printf("[PRODUCER] Gagal parse cache list thread. page=%d limit=%d err=%v", page, limit, err)
	}

	offset := (page - 1) * limit
	threadList, totalForum, err := threadService.threadSQLRepository.GetThreadList(limit, offset)
	if err != nil {
		return response.GetAllThread{}, err
	}

	result := response.GetAllThread{
		ListForum:  threadList,
		TotalForum: totalForum,
	}

	serializedValue, err := json.Marshal(result)
	if err != nil {
		log.Printf("[PRODUCER] Gagal serialize cache list thread. page=%d limit=%d err=%v", page, limit, err)
		return result, nil
	}

	err = threadService.threadListCacheRepository.SetHashFieldWithTTL(
		cacheKey,
		cacheField,
		string(serializedValue),
		time.Duration(model.DefaultThreadListCacheTTLSec)*time.Second,
	)
	if err != nil {
		log.Printf("[PRODUCER] Gagal simpan cache list thread. page=%d limit=%d err=%v", page, limit, err)
	} else {
		log.Printf("[PRODUCER] List thread berhasil disimpan ke cache. page=%d limit=%d", page, limit)
	}

	return result, nil
}

func (threadService ThreadService) GetThreadDetail(threadID int64) (response.ThreadDetail, error) {
	cacheField := buildThreadDetailCacheField(threadID)
	cachedThreadDetailValue, isCacheHit, err := threadService.threadDetailCacheRepository.GetHashField(
		model.CacheThreadDetailHashKey,
		cacheField,
	)
	if err != nil {
		log.Printf("[PRODUCER] Gagal baca cache detail thread. thread_id=%d err=%v", threadID, err)
	}
	if isCacheHit {
		var cachedThreadDetail response.ThreadDetail
		if err = json.Unmarshal([]byte(cachedThreadDetailValue), &cachedThreadDetail); err == nil {
			log.Printf("[PRODUCER] Detail thread diambil dari cache. thread_id=%d", threadID)
			return cachedThreadDetail, nil
		}
		log.Printf("[PRODUCER] Gagal parse cache detail thread. thread_id=%d err=%v", threadID, err)
	}

	threadDetail, err := threadService.threadSQLRepository.GetThreadDetail(threadID)
	if err != nil {
		return response.ThreadDetail{}, err
	}

	commentRows, err := threadService.threadSQLRepository.GetThreadCommentRows(threadID)
	if err != nil {
		return response.ThreadDetail{}, err
	}
	threadDetail.CommentList = buildCommentTree(commentRows)

	serializedValue, err := json.Marshal(threadDetail)
	if err != nil {
		log.Printf("[PRODUCER] Gagal serialize cache detail thread. thread_id=%d err=%v", threadID, err)
		return threadDetail, nil
	}

	err = threadService.threadDetailCacheRepository.SetHashFieldWithTTL(
		model.CacheThreadDetailHashKey,
		cacheField,
		string(serializedValue),
		time.Duration(model.DefaultThreadDetailCacheTTLSec)*time.Second,
	)
	if err != nil {
		log.Printf("[PRODUCER] Gagal simpan cache detail thread. thread_id=%d err=%v", threadID, err)
	} else {
		log.Printf("[PRODUCER] Detail thread berhasil disimpan ke cache. thread_id=%d", threadID)
	}

	return threadDetail, nil
}

func (threadService ThreadService) CreateThread(
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

	err = threadService.publishEventByName(model.EventThreadCreated, threadCreatedEvent.RequestID, threadCreatedEvent)
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

func (threadService ThreadService) CreateComment(
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

	err = threadService.publishEventByName(model.EventCommentCreated, commentCreatedEvent.RequestID, commentCreatedEvent)
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

func (threadService ThreadService) GetThreadLiked() ([]response.LikedThreadData, error) {
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

	messageData, err := threadService.publishRPCEventByName(
		model.EventThreadGetLiked,
		getThreadLikedEvent.RequestID,
		getThreadLikedEvent,
	)
	if err != nil {
		log.Printf("[PRODUCER] Publish event thread.get.liked gagal. request_id=%s err=%v", requestID, err)
		return nil, err
	}

	var likedThreadList []response.LikedThreadData
	if err = json.Unmarshal(messageData, &likedThreadList); err == nil {
		return likedThreadList, nil
	}

	var wrappedResponse struct {
		Data  []response.LikedThreadData `json:"data"`
		Error string                     `json:"error,omitempty"`
	}
	if err = json.Unmarshal(messageData, &wrappedResponse); err == nil {
		if wrappedResponse.Error != "" {
			return nil, fmt.Errorf("%w: %s", model.ErrConsumeEvent, wrappedResponse.Error)
		}
		if wrappedResponse.Data != nil {
			return wrappedResponse.Data, nil
		}
	}

	return nil, fmt.Errorf("%w: invalid rpc response body", model.ErrConsumeEvent)
}

func (threadService ThreadService) InsertLikeThread(threadID int64, likedBy int64) error {
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

	err = threadService.publishEventByName(model.EventThreadLiked, threadLikedEvent.RequestID, threadLikedEvent)
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

func (threadService ThreadService) publishRPCEventByName(
	eventName string,
	requestID string,
	eventPayloadData any,
) ([]byte, error) {
	routingKey, err := messaging.RoutingKeyForEvent(eventName)
	if err != nil {
		return nil, err
	}

	requestContext, cancel := context.WithTimeout(context.Background(), defaultRPCWaitTimeout)
	defer cancel()

	eventPayload, err := json.Marshal(eventPayloadData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", model.ErrPublishEvent, err)
	}

	log.Printf(
		"[PRODUCER] Mulai kirim event %s ke RabbitMQ. request_id=%s exchange=%s routing_key=%s",
		eventName,
		requestID,
		threadService.rabbitRPCClient.ExchangeName,
		routingKey,
	)

	consumerTag := fmt.Sprintf("rpc-%s", requestID)
	msgs, err := threadService.rabbitRPCClient.Channel.Consume(
		threadService.rabbitRPCClient.QueueName,
		consumerTag,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: start consume: %v", model.ErrConsumeEvent, err)
	}
	defer func() {
		_ = threadService.rabbitRPCClient.Channel.Cancel(consumerTag, false)
	}()

	corrID := helper.RandomString(31)

	err = threadService.rabbitRPCClient.Channel.PublishWithContext(
		requestContext,
		threadService.rabbitRPCClient.ExchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       threadService.rabbitRPCClient.QueueName,
			Body:          eventPayload,
		},
	)
	if err != nil {
		log.Printf("[PRODUCER] Gagal kirim event %s ke RabbitMQ. request_id=%s err=%v", eventName, requestID, err)
		return nil, fmt.Errorf("%w: %v", model.ErrPublishEvent, err)
	}

	log.Printf("[PRODUCER] Event %s berhasil dikirim ke RabbitMQ. request_id=%s", eventName, requestID)
	for {
		select {
		case <-requestContext.Done():
			return nil, fmt.Errorf("%w: wait rpc response: %v", model.ErrConsumeEvent, requestContext.Err())
		case delivery, isOpen := <-msgs:
			if !isOpen {
				return nil, fmt.Errorf("%w: rpc response channel closed", model.ErrConsumeEvent)
			}

			if delivery.CorrelationId != "" && delivery.CorrelationId != corrID {
				continue
			}

			log.Printf("[PRODUCER] RPC response %s diterima. request_id=%s", eventName, requestID)
			return delivery.Body, nil
		}
	}
}

func (threadService ThreadService) publishEventByName(
	eventName string,
	requestID string,
	eventPayloadData any,
) error {
	routingKey, err := messaging.RoutingKeyForEvent(eventName)
	if err != nil {
		return err
	}

	requestContext := context.Background()
	eventPayload, err := json.Marshal(eventPayloadData)
	if err != nil {
		return fmt.Errorf("%w: %v", model.ErrPublishEvent, err)
	}

	log.Printf(
		"[PRODUCER] Mulai kirim event %s ke RabbitMQ. request_id=%s exchange=%s routing_key=%s",
		eventName,
		requestID,
		threadService.rabbitPublisher.ExchangeName,
		routingKey,
	)

	err = threadService.rabbitPublisher.Channel.PublishWithContext(
		requestContext,
		threadService.rabbitPublisher.ExchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         eventPayload,
		},
	)
	if err != nil {
		log.Printf("[PRODUCER] Gagal kirim event %s ke RabbitMQ. request_id=%s err=%v", eventName, requestID, err)
		return fmt.Errorf("%w: %v", model.ErrPublishEvent, err)
	}

	log.Printf("[PRODUCER] Event %s berhasil dikirim ke RabbitMQ. request_id=%s", eventName, requestID)
	return nil
}

func buildCommentTree(commentRows []model.ThreadCommentRow) []response.CommentData {
	commentByID := make(map[int64]model.ThreadCommentRow, len(commentRows))
	childByParentID := make(map[int64][]int64)
	topLevelCommentIDs := make([]int64, 0)

	for _, commentRow := range commentRows {
		commentByID[commentRow.ID] = commentRow
		if commentRow.ParentCommentID.Valid {
			parentID := commentRow.ParentCommentID.Int64
			childByParentID[parentID] = append(childByParentID[parentID], commentRow.ID)
			continue
		}

		topLevelCommentIDs = append(topLevelCommentIDs, commentRow.ID)
	}

	var buildReplyTree func(commentID int64) response.CommentData
	buildReplyTree = func(commentID int64) response.CommentData {
		commentRow := commentByID[commentID]
		childCommentIDs := childByParentID[commentID]
		replyList := make([]response.CommentData, 0, len(childCommentIDs))
		for _, childCommentID := range childCommentIDs {
			replyList = append(replyList, buildReplyTree(childCommentID))
		}

		commentData := response.CommentData{
			Comment:    commentRow.Comment,
			CommentBy:  commentRow.CommentBy,
			ReplyList:  replyList,
			TotalReply: commentRow.TotalReply,
		}
		if commentRow.CreatedAt.Valid {
			commentData.CreatedAt = commentRow.CreatedAt.Time
		}

		return commentData
	}

	commentList := make([]response.CommentData, 0, len(topLevelCommentIDs))
	for _, commentID := range topLevelCommentIDs {
		commentList = append(commentList, buildReplyTree(commentID))
	}

	return commentList
}

func buildThreadListCacheKey(limit int) string {
	return fmt.Sprintf("%s:%d", model.CacheThreadListHashKeyPrefix, limit)
}

func buildThreadListCacheField(page int) string {
	return strconv.Itoa(page)
}

func buildThreadDetailCacheField(threadID int64) string {
	return strconv.FormatInt(threadID, 10)
}

func (threadService ThreadService) invalidateThreadListCache() error {
	var cursor uint64
	for {
		cacheKeys, nextCursor, err := threadService.threadListCacheRepository.ScanKeys(
			cursor,
			model.CacheThreadListPattern,
			100,
		)
		if err != nil {
			return err
		}

		if len(cacheKeys) > 0 {
			err = threadService.threadListCacheRepository.DeleteKeys(cacheKeys...)
			if err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

func (threadService ThreadService) invalidateThreadDetailCache(threadID int64) error {
	cacheField := buildThreadDetailCacheField(threadID)
	return threadService.threadDetailCacheRepository.DeleteHashFields(model.CacheThreadDetailHashKey, cacheField)
}
