package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"kaktus-consumer/model"

	"github.com/jackc/pgx/v5/pgconn"
)

func InsertThreadLike(
	requestContext context.Context,
	databaseConnection *sql.DB,
	threadLikedEvent model.ThreadLikedEvent,
) error {
	databaseTransaction, beginTransactionError := databaseConnection.BeginTx(requestContext, nil)
	if beginTransactionError != nil {
		return fmt.Errorf("%w: begin transaction insert thread like: %v", model.ErrConsumeEvent, beginTransactionError)
	}
	defer databaseTransaction.Rollback()

	insertQuery := `
		INSERT INTO thread_likes (thread_id, liked_by)
		VALUES ($1, $2)
		ON CONFLICT (thread_id, liked_by) DO NOTHING
	`

	insertResult, insertError := databaseTransaction.ExecContext(
		requestContext,
		insertQuery,
		threadLikedEvent.ThreadID,
		threadLikedEvent.LikedBy,
	)
	if insertError != nil {
		if isInvalidLikeInsertPayloadError(insertError) {
			return fmt.Errorf("%w: insert thread like: %v", model.ErrInvalidEventPayload, insertError)
		}
		return fmt.Errorf("%w: insert thread like: %v", model.ErrConsumeEvent, insertError)
	}

	insertedRowsCount, insertedRowsError := insertResult.RowsAffected()
	if insertedRowsError != nil {
		return fmt.Errorf("%w: read affected rows insert thread like: %v", model.ErrConsumeEvent, insertedRowsError)
	}

	if insertedRowsCount > 0 {
		incrementTotalLikeQuery := `
			UPDATE threads
			SET total_likes = total_likes + 1, updated_at = NOW()
			WHERE id = $1
		`

		updateResult, updateTotalLikeError := databaseTransaction.ExecContext(
			requestContext,
			incrementTotalLikeQuery,
			threadLikedEvent.ThreadID,
		)
		if updateTotalLikeError != nil {
			return fmt.Errorf("%w: update thread total_likes: %v", model.ErrConsumeEvent, updateTotalLikeError)
		}

		updatedRowsCount, updatedRowsError := updateResult.RowsAffected()
		if updatedRowsError != nil {
			return fmt.Errorf("%w: read affected rows total_likes: %v", model.ErrConsumeEvent, updatedRowsError)
		}
		if updatedRowsCount == 0 {
			return fmt.Errorf("%w: update thread total_likes: thread not found", model.ErrConsumeEvent)
		}
	}

	if commitTransactionError := databaseTransaction.Commit(); commitTransactionError != nil {
		return fmt.Errorf("%w: commit transaction insert thread like: %v", model.ErrConsumeEvent, commitTransactionError)
	}

	return nil
}

func isInvalidLikeInsertPayloadError(insertError error) bool {
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
