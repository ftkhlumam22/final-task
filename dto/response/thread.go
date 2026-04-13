package response

import "time"

type ThreadData struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type GetAllThread struct {
	ListForum  []ThreadData `json:"list_forum"`
	TotalForum int          `json:"total_forum"`
}

type ThreadDetail struct {
	ID            int64         `json:"id"`
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	CreatedBy     string        `json:"created_by"`
	CreatedAt     time.Time     `json:"created_at"`
	TotalLikes    int           `json:"total_likes"`
	TotalComments int           `json:"total_comments"`
	CommentList   []CommentData `json:"comment_list"`
}
