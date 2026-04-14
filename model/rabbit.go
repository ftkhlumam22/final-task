package model

type RabbitConfig struct {
	URL      string
	MaxRetry int
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
	ParentCommentID *int64 `json:"parent_comment_id"`
}

type ThreadLikedEvent struct {
	Event     string `json:"event"`
	RequestID string `json:"request_id"`
	ThreadID  int64  `json:"thread_id"`
	LikedBy   int64  `json:"liked_by"`
}

type EventEnvelope struct {
	Event     string `json:"event"`
	RequestID string `json:"request_id"`
}

type ThreadGetLikedRequest struct {
	Event     string         `json:"event"`
	RequestID string         `json:"request_id"`
	Body      map[string]any `json:"body"`
}

type ThreadLikedListItem struct {
	Title   string `json:"title"`
	LikedBy string `json:"liked_by"`
}

type ThreadGetLikedSuccessResponse struct {
	Data []ThreadLikedListItem `json:"data"`
}

type ThreadGetLikedErrorResponse struct {
	Error string `json:"error"`
}
