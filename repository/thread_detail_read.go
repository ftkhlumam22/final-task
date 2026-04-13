package repository

import (
	"context"
	"database/sql"
	"errors"
	"final-task/dto/response"
	"final-task/model"
)

type threadCommentRecord struct {
	ID              int64
	Comment         string
	TotalReply      int
	CommentBy       string
	CreatedAt       sql.NullTime
	ParentCommentID sql.NullInt64
}

func GetThreadDetail(
	requestContext context.Context,
	databaseConnection *sql.DB,
	threadID int64,
) (response.ThreadDetail, error) {
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
	getDetailError := databaseConnection.QueryRowContext(requestContext, threadDetailQuery, threadID).Scan(
		&threadDetail.ID,
		&threadDetail.Title,
		&threadDetail.Description,
		&threadDetail.CreatedBy,
		&threadDetail.CreatedAt,
		&threadDetail.TotalLikes,
		&threadDetail.TotalComments,
	)
	if getDetailError != nil {
		if errors.Is(getDetailError, sql.ErrNoRows) {
			return response.ThreadDetail{}, model.ErrThreadNotFound
		}
		return response.ThreadDetail{}, getDetailError
	}

	threadDetail.CommentList = make([]response.CommentData, 0)
	return threadDetail, nil
}

func GetThreadCommentTree(
	requestContext context.Context,
	databaseConnection *sql.DB,
	threadID int64,
) ([]response.CommentData, error) {
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

	rows, queryError := databaseConnection.QueryContext(requestContext, commentListQuery, threadID)
	if queryError != nil {
		return nil, queryError
	}
	defer rows.Close()

	commentByID := make(map[int64]threadCommentRecord)
	childByParentID := make(map[int64][]int64)
	topLevelCommentIDs := make([]int64, 0)
	for rows.Next() {
		var commentRecord threadCommentRecord
		scanError := rows.Scan(
			&commentRecord.ID,
			&commentRecord.Comment,
			&commentRecord.TotalReply,
			&commentRecord.CommentBy,
			&commentRecord.CreatedAt,
			&commentRecord.ParentCommentID,
		)
		if scanError != nil {
			return nil, scanError
		}

		commentByID[commentRecord.ID] = commentRecord

		if commentRecord.ParentCommentID.Valid {
			parentID := commentRecord.ParentCommentID.Int64
			childByParentID[parentID] = append(childByParentID[parentID], commentRecord.ID)
			continue
		}

		topLevelCommentIDs = append(topLevelCommentIDs, commentRecord.ID)
	}
	if rowsError := rows.Err(); rowsError != nil {
		return nil, rowsError
	}

	var buildReplyTree func(commentID int64) response.CommentData
	buildReplyTree = func(commentID int64) response.CommentData {
		commentRecord := commentByID[commentID]
		childCommentIDs := childByParentID[commentID]
		replyList := make([]response.CommentData, 0, len(childCommentIDs))
		for _, childCommentID := range childCommentIDs {
			replyList = append(replyList, buildReplyTree(childCommentID))
		}

		commentData := response.CommentData{
			Comment:    commentRecord.Comment,
			CommentBy:  commentRecord.CommentBy,
			ReplyList:  replyList,
			TotalReply: commentRecord.TotalReply,
		}
		if commentRecord.CreatedAt.Valid {
			commentData.CreatedAt = commentRecord.CreatedAt.Time
		}

		return commentData
	}

	commentList := make([]response.CommentData, 0, len(topLevelCommentIDs))
	for _, commentID := range topLevelCommentIDs {
		commentList = append(commentList, buildReplyTree(commentID))
	}

	return commentList, nil
}
