package thread

import (
	"context"
	"time"

	"kaktus-consumer/model"
	"kaktus-consumer/repository"
)

type Service interface {
	HandleThreadCreatedMessage(requestContext context.Context, threadCreatedEvent model.ThreadCreatedEvent) ConsumeDecision
	HandleCommentCreatedMessage(requestContext context.Context, commentCreatedEvent model.CommentCreatedEvent) ConsumeDecision
	HandleThreadLikedMessage(requestContext context.Context, threadLikedEvent model.ThreadLikedEvent) ConsumeDecision
	HandleThreadGetLikedMessage(
		requestContext context.Context,
		threadGetLikedRequest model.ThreadGetLikedRequest,
		replyToQueue string,
		correlationID string,
	) ConsumeDecision
}

type RPCPublisherRepository interface {
	PublishRPCResponse(
		requestContext context.Context,
		replyToQueue string,
		correlationID string,
		responsePayload interface{},
	) error
	Close() error
}

type Dependency struct {
	ThreadRepository    repository.ThreadRepository
	PublisherRepository RPCPublisherRepository
	MaxRetry            int
	RetryDelay          time.Duration
}

type ConsumeDecision struct {
	EventName  string
	RequestID  string
	ShouldAck  bool
	Requeue    bool
	NackReason string
	Err        error
}

type service struct {
	threadRepository    repository.ThreadRepository
	publisherRepository RPCPublisherRepository
	maxRetry            int
	retryDelay          time.Duration
}
