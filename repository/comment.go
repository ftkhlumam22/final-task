package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"kaktus-consumer/model"

	"github.com/jackc/pgx/v5/pgconn"
)

func InsertComment(
	requestContext context.Context,
	databaseConnection *sql.DB,
	commentCreatedEvent model.CommentCreatedEvent,
) error {
	databaseTransaction, beginTransactionError := databaseConnection.BeginTx(requestContext, nil)
	if beginTransactionError != nil {
		return fmt.Errorf("%w: begin transaction insert comment: %v", model.ErrConsumeEvent, beginTransactionError)
	}
	defer databaseTransaction.Rollback()

	insertQuery := `
		INSERT INTO comments (thread_id, comment, created_by, parent_comment_id)
		VALUES ($1, $2, $3, $4)
	`

	_, insertError := databaseTransaction.ExecContext(
		requestContext,
		insertQuery,
		commentCreatedEvent.ThreadID,
		commentCreatedEvent.Comment,
		commentCreatedEvent.CreatedBy,
		commentCreatedEvent.ParentCommentID,
	)
	if insertError != nil {
		if isInvalidCommentInsertPayloadError(insertError) {
			return fmt.Errorf("%w: insert comment: %v", model.ErrInvalidEventPayload, insertError)
		}
		return fmt.Errorf("%w: insert comment: %v", model.ErrConsumeEvent, insertError)
	}

	if commentCreatedEvent.ParentCommentID != nil {
		incrementParentReplyQuery := `
			UPDATE comments
			SET total_reply = total_reply + 1, updated_at = NOW()
			WHERE id = $1 AND thread_id = $2
		`

		updateParentReplyResult, updateParentReplyError := databaseTransaction.ExecContext(
			requestContext,
			incrementParentReplyQuery,
			*commentCreatedEvent.ParentCommentID,
			commentCreatedEvent.ThreadID,
		)
		if updateParentReplyError != nil {
			return fmt.Errorf("%w: update parent comment total_reply: %v", model.ErrConsumeEvent, updateParentReplyError)
		}

		updatedParentRowsCount, updatedParentRowsError := updateParentReplyResult.RowsAffected()
		if updatedParentRowsError != nil {
			return fmt.Errorf("%w: read affected rows parent total_reply: %v", model.ErrConsumeEvent, updatedParentRowsError)
		}
		if updatedParentRowsCount == 0 {
			return fmt.Errorf("%w: update parent comment total_reply: parent comment not found", model.ErrInvalidEventPayload)
		}
	}

	if commentCreatedEvent.ParentCommentID == nil {
		incrementTotalCommentQuery := `
		UPDATE threads
		SET total_comment = total_comment + 1, updated_at = NOW()
		WHERE id = $1
	`

		updateResult, updateTotalCommentError := databaseTransaction.ExecContext(
			requestContext,
			incrementTotalCommentQuery,
			commentCreatedEvent.ThreadID,
		)
		if updateTotalCommentError != nil {
			return fmt.Errorf("%w: update thread total_comment: %v", model.ErrConsumeEvent, updateTotalCommentError)
		}

		updatedRowsCount, updatedRowsError := updateResult.RowsAffected()
		if updatedRowsError != nil {
			return fmt.Errorf("%w: read affected rows total_comment: %v", model.ErrConsumeEvent, updatedRowsError)
		}
		if updatedRowsCount == 0 {
			return fmt.Errorf("%w: update thread total_comment: thread not found", model.ErrConsumeEvent)
		}
	}

	if commitTransactionError := databaseTransaction.Commit(); commitTransactionError != nil {
		return fmt.Errorf("%w: commit transaction insert comment: %v", model.ErrConsumeEvent, commitTransactionError)
	}

	return nil
}

func isInvalidCommentInsertPayloadError(insertError error) bool {
	var postgresError *pgconn.PgError
	if !errors.As(insertError, &postgresError) {
		return false
	}

	switch postgresError.Code {
	case "23503", // foreign_key_violation
		"23502", // not_null_violation
		"22001", // string_data_right_truncation
		"22P02": // invalid_text_representation
		return true
	default:
		return false
	}
}
