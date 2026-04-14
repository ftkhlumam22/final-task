package messaging

import (
	"fmt"

	"final-task/model"
)

const (
	ExchangeName = "kaktus.events"
	ExchangeType = "topic"
)

const (
	threadCreatedRoutingKey  = "thread.created"
	commentCreatedRoutingKey = "comment.created"
	threadLikedRoutingKey    = "thread.liked"
	threadGetLikedRoutingKey = "thread.get.liked"
)

func RoutingKeyForEvent(eventName string) (string, error) {
	switch eventName {
	case model.EventThreadCreated:
		return threadCreatedRoutingKey, nil
	case model.EventCommentCreated:
		return commentCreatedRoutingKey, nil
	case model.EventThreadLiked:
		return threadLikedRoutingKey, nil
	case model.EventThreadGetLiked:
		return threadGetLikedRoutingKey, nil
	default:
		return "", fmt.Errorf("%w: unknown event %q", model.ErrPublishEvent, eventName)
	}
}
