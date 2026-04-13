package model

import amqp "github.com/rabbitmq/amqp091-go"

type RabbitConfig struct {
	URL                      string
	ExchangeName             string
	ExchangeType             string
	ThreadCreatedRoutingKey  string
	CommentCreatedRoutingKey string
	ThreadLikedRoutingKey    string
}

type RabbitPublisher struct {
	Connection               *amqp.Connection
	Channel                  *amqp.Channel
	ExchangeName             string
	ThreadCreatedRoutingKey  string
	CommentCreatedRoutingKey string
	ThreadLikedRoutingKey    string
}

type ThreadCreatedEvent struct {
	Event       string `json:"event"`
	RequestID   string `json:"request_id"`
	CreatedBy   int64  `json:"created_by"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type CommentCreatedEvent struct {
	Event           string `json:"event"`
	RequestID       string `json:"request_id"`
	ThreadID        int64  `json:"thread_id"`
	CreatedBy       int64  `json:"created_by"`
	Comment         string `json:"comment"`
	ParentCommentID *int64 `json:"parent_comment_id,omitempty"`
}

type ThreadLikedEvent struct {
	Event     string `json:"event"`
	RequestID string `json:"request_id"`
	ThreadID  int64  `json:"thread_id"`
	LikedBy   int64  `json:"liked_by"`
}
