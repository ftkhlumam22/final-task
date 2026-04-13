package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"kaktus-consumer/model"

	"github.com/jackc/pgx/v5/pgconn"
)

func InsertThread(
	requestContext context.Context,
	databaseConnection *sql.DB,
	threadCreatedEvent model.ThreadCreatedEvent,
) (model.Thread, error) {
	insertQuery := `
		INSERT INTO threads (title, description, created_by)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	insertedThread := model.Thread{
		Title:       threadCreatedEvent.Title,
		Description: threadCreatedEvent.Description,
		CreatedBy:   threadCreatedEvent.CreatedBy,
	}

	insertError := databaseConnection.QueryRowContext(
		requestContext,
		insertQuery,
		threadCreatedEvent.Title,
		threadCreatedEvent.Description,
		threadCreatedEvent.CreatedBy,
	).Scan(&insertedThread.ID, &insertedThread.CreatedAt)
	if insertError != nil {
		if isInvalidThreadInsertPayloadError(insertError) {
			return model.Thread{}, fmt.Errorf("%w: insert thread: %v", model.ErrInvalidEventPayload, insertError)
		}
		return model.Thread{}, fmt.Errorf("%w: insert thread: %v", model.ErrConsumeEvent, insertError)
	}

	return insertedThread, nil
}

func isInvalidThreadInsertPayloadError(insertError error) bool {
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
