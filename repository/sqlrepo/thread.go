package sqlrepo

import (
	"context"

	"kaktus-consumer/model"
)

func (repository *SQLRepository) InsertThread(
	requestContext context.Context,
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

	err := repository.databaseConnection.QueryRowContext(
		requestContext,
		insertQuery,
		threadCreatedEvent.Title,
		threadCreatedEvent.Description,
		threadCreatedEvent.CreatedBy,
	).Scan(&insertedThread.ID, &insertedThread.CreatedAt)
	if err != nil {
		return model.Thread{}, wrapSQLError("insert thread", err)
	}

	return insertedThread, nil
}
