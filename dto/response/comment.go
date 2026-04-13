package response

import "time"

type CommentData struct {
	Comment    string        `json:"comment"`
	CommentBy  string        `json:"comment_by"`
	CreatedAt  time.Time     `json:"created_at"`
	ReplyList  []CommentData `json:"reply_list"`
	TotalReply int           `json:"total_reply"`
}
