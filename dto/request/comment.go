package request

type Comment struct {
	Comment         string `json:"comment"`
	ThreadID        int64  `json:"thread_id"`
	ParentCommentID *int64 `json:"parent_comment_id,omitempty"`
}
