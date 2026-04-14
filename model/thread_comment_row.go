package model

import "database/sql"

type ThreadCommentRow struct {
	ID              int64
	Comment         string
	TotalReply      int
	CommentBy       string
	CreatedAt       sql.NullTime
	ParentCommentID sql.NullInt64
}
