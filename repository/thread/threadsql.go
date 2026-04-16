package threadrepo

import (
	"context"
	"database/sql"
	"errors"

	"final-task/dto/response"
	"final-task/model"
)

func (threadSQLRepository ThreadSQLRepository) GetThreadList(limit int, offset int) ([]response.ThreadData, int, error) {
	threadListQuery := `
		SELECT
			threads.id,
			threads.title,
			threads.description,
			users.name AS created_by,
			threads.created_at
		FROM threads
		JOIN users ON users.id = threads.created_by
		ORDER BY threads.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := threadSQLRepository.db.QueryContext(context.Background(), threadListQuery, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	threadList := make([]response.ThreadData, 0)
	for rows.Next() {
		var threadData response.ThreadData
		err = rows.Scan(
			&threadData.ID,
			&threadData.Title,
			&threadData.Description,
			&threadData.CreatedBy,
			&threadData.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		threadList = append(threadList, threadData)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	var totalForum int
	err = threadSQLRepository.db.QueryRowContext(
		context.Background(),
		`SELECT COUNT(1) FROM threads`,
	).Scan(&totalForum)
	if err != nil {
		return nil, 0, err
	}

	return threadList, totalForum, nil
}

func (threadSQLRepository ThreadSQLRepository) GetThreadDetail(threadID int64) (response.ThreadDetail, error) {
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
	err := threadSQLRepository.db.QueryRowContext(context.Background(), threadDetailQuery, threadID).Scan(
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

func (threadSQLRepository ThreadSQLRepository) GetThreadCommentRows(threadID int64) ([]model.ThreadCommentRow, error) {
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

	rows, err := threadSQLRepository.db.QueryContext(context.Background(), commentListQuery, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	commentRows := make([]model.ThreadCommentRow, 0)
	for rows.Next() {
		var commentRow model.ThreadCommentRow
		err = rows.Scan(
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
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return commentRows, nil
}
