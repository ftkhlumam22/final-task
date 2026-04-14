package sqlrepo

import (
	"context"
	"fmt"

	"kaktus-consumer/model"
)

func (repository *sqlTransactionRepository) InsertComment(
	requestContext context.Context,
	commentCreatedEvent model.CommentCreatedEvent,
) error {
	insertQuery := `
		INSERT INTO comments (thread_id, comment, created_by, parent_comment_id)
		VALUES ($1, $2, $3, $4)
	`

	_, err := repository.transaction.ExecContext(
		requestContext,
		insertQuery,
		commentCreatedEvent.ThreadID,
		commentCreatedEvent.Comment,
		commentCreatedEvent.CreatedBy,
		commentCreatedEvent.ParentCommentID,
	)
	if err != nil {
		return wrapSQLError("insert comment", err)
	}

	return nil
}

func (repository *sqlTransactionRepository) IncrementParentCommentReply(
	requestContext context.Context,
	parentCommentID int64,
	threadID int64,
) (int64, error) {
	updateQuery := `
		UPDATE comments
		SET total_reply = total_reply + 1, updated_at = NOW()
		WHERE id = $1 AND thread_id = $2
	`

	updateResult, err := repository.transaction.ExecContext(
		requestContext,
		updateQuery,
		parentCommentID,
		threadID,
	)
	if err != nil {
		return 0, fmt.Errorf("%w: update parent comment total_reply: %v", model.ErrConsumeEvent, err)
	}

	rowsAffected, err := updateResult.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%w: read affected rows parent total_reply: %v", model.ErrConsumeEvent, err)
	}

	return rowsAffected, nil
}

func (repository *sqlTransactionRepository) IncrementThreadTotalComment(
	requestContext context.Context,
	threadID int64,
) (int64, error) {
	updateQuery := `
		UPDATE threads
		SET total_comment = total_comment + 1, updated_at = NOW()
		WHERE id = $1
	`

	updateResult, err := repository.transaction.ExecContext(
		requestContext,
		updateQuery,
		threadID,
	)
	if err != nil {
		return 0, fmt.Errorf("%w: update thread total_comment: %v", model.ErrConsumeEvent, err)
	}

	rowsAffected, err := updateResult.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%w: read affected rows total_comment: %v", model.ErrConsumeEvent, err)
	}

	return rowsAffected, nil
}
