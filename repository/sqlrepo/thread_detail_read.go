package sqlrepo

import (
	"context"
	"database/sql"
	"errors"
	"final-task/dto/response"
	"final-task/model"
)

func (threadReadRepository *ThreadReadRepository) GetThreadDetail(threadID int64) (response.ThreadDetail, error) {
	threadDetailQuery := `
		SELECT
			threads.id,
			threads.title,
			threads.description,
			users.name AS created_by,
			threads.created_at,
			threads.total_likes,
			threads.total_comment
		FROM threads
		JOIN users ON users.id = threads.created_by
		WHERE threads.id = $1
	`

	var threadDetail response.ThreadDetail
	err := threadReadRepository.db.QueryRowContext(context.Background(), threadDetailQuery, threadID).Scan(
		&threadDetail.ID,
		&threadDetail.Title,
		&threadDetail.Description,
		&threadDetail.CreatedBy,
		&threadDetail.CreatedAt,
		&threadDetail.TotalLikes,
		&threadDetail.TotalComments,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response.ThreadDetail{}, model.ErrThreadNotFound
		}
		return response.ThreadDetail{}, err
	}

	threadDetail.CommentList = make([]response.CommentData, 0)
	return threadDetail, nil
}

func (threadReadRepository *ThreadReadRepository) GetThreadCommentRows(threadID int64) ([]ThreadCommentRow, error) {
	commentListQuery := `
		SELECT
			comments.id,
			comments.comment,
			comments.total_reply,
			users.name AS comment_by,
			comments.created_at,
			comments.parent_comment_id
		FROM comments
		JOIN users ON users.id = comments.created_by
		WHERE comments.thread_id = $1
		ORDER BY comments.created_at ASC, comments.id ASC
	`

	rows, err := threadReadRepository.db.QueryContext(context.Background(), commentListQuery, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	commentRows := make([]ThreadCommentRow, 0)
	for rows.Next() {
		var commentRow ThreadCommentRow
		err := rows.Scan(
			&commentRow.ID,
			&commentRow.Comment,
			&commentRow.TotalReply,
			&commentRow.CommentBy,
			&commentRow.CreatedAt,
			&commentRow.ParentCommentID,
		)
		if err != nil {
			return nil, err
		}

		commentRows = append(commentRows, commentRow)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return commentRows, nil
}
