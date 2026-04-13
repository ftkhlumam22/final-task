package module

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"strings"
	"unicode/utf8"

	"kaktus-consumer/model"
	"kaktus-consumer/repository"

	"github.com/redis/go-redis/v9"
)

func ParseEventEnvelope(messageBody []byte) (model.EventEnvelope, error) {
	var eventEnvelope model.EventEnvelope
	if unmarshalError := json.Unmarshal(messageBody, &eventEnvelope); unmarshalError != nil {
		return model.EventEnvelope{}, model.ErrInvalidEventPayload
	}

	eventEnvelope.Event = strings.TrimSpace(eventEnvelope.Event)
	eventEnvelope.RequestID = strings.TrimSpace(eventEnvelope.RequestID)
	if eventEnvelope.Event == "" || eventEnvelope.RequestID == "" {
		return model.EventEnvelope{}, model.ErrInvalidEventPayload
	}

	return eventEnvelope, nil
}

func ParseThreadCreatedEvent(messageBody []byte) (model.ThreadCreatedEvent, error) {
	var threadCreatedEvent model.ThreadCreatedEvent
	if unmarshalError := json.Unmarshal(messageBody, &threadCreatedEvent); unmarshalError != nil {
		return model.ThreadCreatedEvent{}, model.ErrInvalidEventPayload
	}
	return threadCreatedEvent, nil
}

func ParseCommentCreatedEvent(messageBody []byte) (model.CommentCreatedEvent, error) {
	var commentCreatedEvent model.CommentCreatedEvent
	if unmarshalError := json.Unmarshal(messageBody, &commentCreatedEvent); unmarshalError != nil {
		return model.CommentCreatedEvent{}, model.ErrInvalidEventPayload
	}
	return commentCreatedEvent, nil
}

func ParseThreadLikedEvent(messageBody []byte) (model.ThreadLikedEvent, error) {
	var threadLikedEvent model.ThreadLikedEvent
	if unmarshalError := json.Unmarshal(messageBody, &threadLikedEvent); unmarshalError != nil {
		return model.ThreadLikedEvent{}, model.ErrInvalidEventPayload
	}
	return threadLikedEvent, nil
}

func ProcessThreadCreatedEvent(
	requestContext context.Context,
	databaseConnection *sql.DB,
	redisClient *redis.Client,
	threadCreatedEvent model.ThreadCreatedEvent,
) error {
	threadCreatedEvent.Event = strings.TrimSpace(threadCreatedEvent.Event)
	threadCreatedEvent.RequestID = strings.TrimSpace(threadCreatedEvent.RequestID)
	threadCreatedEvent.Title = strings.TrimSpace(threadCreatedEvent.Title)
	threadCreatedEvent.Description = strings.TrimSpace(threadCreatedEvent.Description)

	if validateError := ValidateThreadCreatedEvent(threadCreatedEvent); validateError != nil {
		log.Printf("[consumer][thread.created] validasi payload gagal. request_id=%s", threadCreatedEvent.RequestID)
		return validateError
	}
	log.Printf("[consumer][thread.created] validasi payload sukses. request_id=%s", threadCreatedEvent.RequestID)

	if _, insertError := repository.InsertThread(requestContext, databaseConnection, threadCreatedEvent); insertError != nil {
		log.Printf("[consumer][thread.created] write DB gagal. request_id=%s error=%v", threadCreatedEvent.RequestID, insertError)
		return insertError
	}
	log.Printf("[consumer][thread.created] write DB sukses. request_id=%s", threadCreatedEvent.RequestID)

	if invalidateError := repository.InvalidateThreadListCache(requestContext, redisClient); invalidateError != nil {
		log.Printf("[consumer][thread.created] invalidate cache gagal. request_id=%s error=%v", threadCreatedEvent.RequestID, invalidateError)
		return invalidateError
	}
	log.Printf("[consumer][thread.created] cache list invalidated. request_id=%s", threadCreatedEvent.RequestID)

	return nil
}

func ProcessCommentCreatedEvent(
	requestContext context.Context,
	databaseConnection *sql.DB,
	redisClient *redis.Client,
	commentCreatedEvent model.CommentCreatedEvent,
) error {
	commentCreatedEvent.Event = strings.TrimSpace(commentCreatedEvent.Event)
	commentCreatedEvent.RequestID = strings.TrimSpace(commentCreatedEvent.RequestID)
	commentCreatedEvent.Comment = strings.TrimSpace(commentCreatedEvent.Comment)

	if validateError := ValidateCommentCreatedEvent(commentCreatedEvent); validateError != nil {
		log.Printf("[consumer][comment.created] validasi payload gagal. request_id=%s", commentCreatedEvent.RequestID)
		return validateError
	}
	log.Printf("[consumer][comment.created] validasi payload sukses. request_id=%s", commentCreatedEvent.RequestID)

	if insertError := repository.InsertComment(requestContext, databaseConnection, commentCreatedEvent); insertError != nil {
		log.Printf("[consumer][comment.created] write DB gagal. request_id=%s error=%v", commentCreatedEvent.RequestID, insertError)
		return insertError
	}
	log.Printf("[consumer][comment.created] write DB sukses. request_id=%s", commentCreatedEvent.RequestID)

	if invalidateError := repository.InvalidateThreadDetailCache(requestContext, redisClient, commentCreatedEvent.ThreadID); invalidateError != nil {
		log.Printf("[consumer][comment.created] invalidate cache gagal. request_id=%s error=%v", commentCreatedEvent.RequestID, invalidateError)
		return invalidateError
	}
	log.Printf("[consumer][comment.created] cache detail invalidated. request_id=%s thread_id=%d", commentCreatedEvent.RequestID, commentCreatedEvent.ThreadID)

	return nil
}

func ProcessThreadLikedEvent(
	requestContext context.Context,
	databaseConnection *sql.DB,
	redisClient *redis.Client,
	threadLikedEvent model.ThreadLikedEvent,
) error {
	threadLikedEvent.Event = strings.TrimSpace(threadLikedEvent.Event)
	threadLikedEvent.RequestID = strings.TrimSpace(threadLikedEvent.RequestID)

	if validateError := ValidateThreadLikedEvent(threadLikedEvent); validateError != nil {
		log.Printf("[consumer][thread.liked] validasi payload gagal. request_id=%s", threadLikedEvent.RequestID)
		return validateError
	}
	log.Printf("[consumer][thread.liked] validasi payload sukses. request_id=%s", threadLikedEvent.RequestID)

	if insertError := repository.InsertThreadLike(requestContext, databaseConnection, threadLikedEvent); insertError != nil {
		log.Printf("[consumer][thread.liked] write DB gagal. request_id=%s error=%v", threadLikedEvent.RequestID, insertError)
		return insertError
	}
	log.Printf("[consumer][thread.liked] write DB sukses. request_id=%s", threadLikedEvent.RequestID)

	if invalidateError := repository.InvalidateThreadDetailCache(requestContext, redisClient, threadLikedEvent.ThreadID); invalidateError != nil {
		log.Printf("[consumer][thread.liked] invalidate cache gagal. request_id=%s error=%v", threadLikedEvent.RequestID, invalidateError)
		return invalidateError
	}
	log.Printf("[consumer][thread.liked] cache detail invalidated. request_id=%s thread_id=%d", threadLikedEvent.RequestID, threadLikedEvent.ThreadID)

	return nil
}

func ValidateCommentCreatedEvent(commentCreatedEvent model.CommentCreatedEvent) error {
	commentCreatedEvent.Event = strings.TrimSpace(commentCreatedEvent.Event)
	commentCreatedEvent.RequestID = strings.TrimSpace(commentCreatedEvent.RequestID)
	commentCreatedEvent.Comment = strings.TrimSpace(commentCreatedEvent.Comment)

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

func ValidateThreadLikedEvent(threadLikedEvent model.ThreadLikedEvent) error {
	threadLikedEvent.Event = strings.TrimSpace(threadLikedEvent.Event)
	threadLikedEvent.RequestID = strings.TrimSpace(threadLikedEvent.RequestID)

	if threadLikedEvent.Event != model.EventThreadLiked ||
		threadLikedEvent.RequestID == "" ||
		threadLikedEvent.ThreadID <= 0 ||
		threadLikedEvent.LikedBy <= 0 {
		return model.ErrInvalidEventPayload
	}

	return nil
}
