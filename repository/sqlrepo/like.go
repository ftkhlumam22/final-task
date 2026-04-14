package sqlrepo

import (
	"context"
	"fmt"

	"kaktus-consumer/model"
)

func (repository *sqlTransactionRepository) InsertThreadLike(
	requestContext context.Context,
	threadLikedEvent model.ThreadLikedEvent,
) (int64, error) {
	insertQuery := `
		INSERT INTO thread_likes (thread_id, liked_by)
		VALUES ($1, $2)
		ON CONFLICT (thread_id, liked_by) DO NOTHING
	`

	insertResult, err := repository.transaction.ExecContext(
		requestContext,
		insertQuery,
		threadLikedEvent.ThreadID,
		threadLikedEvent.LikedBy,
	)
	if err != nil {
		return 0, wrapSQLError("insert thread like", err)
	}

	rowsAffected, err := insertResult.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%w: read affected rows insert thread like: %v", model.ErrConsumeEvent, err)
	}

	return rowsAffected, nil
}

func (repository *sqlTransactionRepository) IncrementThreadTotalLike(
	requestContext context.Context,
	threadID int64,
) (int64, error) {
	updateQuery := `
		UPDATE threads
		SET total_likes = total_likes + 1, updated_at = NOW()
		WHERE id = $1
	`

	updateResult, err := repository.transaction.ExecContext(
		requestContext,
		updateQuery,
		threadID,
	)
	if err != nil {
		return 0, fmt.Errorf("%w: update thread total_likes: %v", model.ErrConsumeEvent, err)
	}

	rowsAffected, err := updateResult.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%w: read affected rows total_likes: %v", model.ErrConsumeEvent, err)
	}

	return rowsAffected, nil
}

func (repository *SQLRepository) GetLikedThreads(
	requestContext context.Context,
) ([]model.ThreadLikedListItem, error) {
	selectQuery := `
		SELECT
			threads.title,
			thread_likes.liked_by::text AS liked_by
		FROM thread_likes
		INNER JOIN threads ON threads.id = thread_likes.thread_id
		ORDER BY thread_likes.thread_id DESC, thread_likes.liked_by DESC
	`

	queryResult, err := repository.databaseConnection.QueryContext(requestContext, selectQuery)
	if err != nil {
		return nil, fmt.Errorf("%w: get liked threads: %v", model.ErrConsumeEvent, err)
	}
	defer queryResult.Close()

	likedThreads := make([]model.ThreadLikedListItem, 0)
	for queryResult.Next() {
		likedThread := model.ThreadLikedListItem{}
		err = queryResult.Scan(&likedThread.Title, &likedThread.LikedBy)
		if err != nil {
			return nil, fmt.Errorf("%w: scan liked threads: %v", model.ErrConsumeEvent, err)
		}
		likedThreads = append(likedThreads, likedThread)
	}

	err = queryResult.Err()
	if err != nil {
		return nil, fmt.Errorf("%w: iterate liked threads: %v", model.ErrConsumeEvent, err)
	}

	return likedThreads, nil
}
