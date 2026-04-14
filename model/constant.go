package model

const (
	StatusSuccess = "success"
	StatusError   = "error"
)

const (
	EventThreadCreated  = "thread.created"
	EventCommentCreated = "comment.created"
	EventThreadLiked    = "thread.liked"
	EventThreadGetLiked = "thread.get.liked"
)

const (
	ThreadListCachePattern      = "thread:list:limit:*"
	ThreadDetailCacheKeyPattern = "thread:detail:%d"
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
	MessageInternalServerError     = "internal server error"
	MessageInvalidEventPayload     = "invalid event payload"
	MessageFailedConsumeEvent      = "failed to consume event"
	MessageFailedFetchLikedThreads = "failed to fetch liked threads"
	MessageMethodNotAllowed        = "method not allowed"
)
