package model

const (
	StatusSuccess = "success"
	StatusError   = "error"
)

const (
	EventThreadCreated  = "thread.created"
	EventCommentCreated = "comment.created"
	EventThreadLiked    = "thread.liked"
)

const (
	ThreadListCachePattern        = "thread:list:limit:*"
	ThreadDetailCacheKeyPattern   = "thread:detail:%d"
	ThreadDetailCacheFieldPayload = "payload"
)

const (
	RabbitHeaderRequestID = "x-request-id"
)

const (
	ConsumerRetryDelaySeconds = 2
)

const (
	MaxThreadTitleLength       = 200
	MaxThreadDescriptionLength = 5000
	MaxCommentLength           = 5000
)

const (
	MessageInternalServerError = "internal server error"
	MessageInvalidEventPayload = "invalid event payload"
	MessageFailedConsumeEvent  = "failed to consume event"
	MessageMethodNotAllowed    = "method not allowed"
)
