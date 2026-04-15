package thread

import (
	"context"
	"errors"
	"fmt"

	"kaktus-consumer/helper"
	"kaktus-consumer/model"

	"github.com/jackc/pgx/v5/pgconn"
)

const cacheScanBatchSize int64 = 100

func (repository *repository) InsertThread(
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

func (repository *repository) WithTransaction(
	requestContext context.Context,
	operation func(transactionRepository TransactionRepository) error,
) error {
	transaction, err := repository.databaseConnection.BeginTx(requestContext, nil)
	if err != nil {
		return fmt.Errorf("%w: begin transaction: %v", model.ErrConsumeEvent, err)
	}
	defer transaction.Rollback()

	if err = operation(&sqlTransactionRepository{transaction: transaction}); err != nil {
		return err
	}

	if err = transaction.Commit(); err != nil {
		return fmt.Errorf("%w: commit transaction: %v", model.ErrConsumeEvent, err)
	}

	return nil
}

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

func (repository *repository) GetLikedThreads(
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

func (repository *repository) InvalidateThreadListCache(requestContext context.Context) error {
	cacheKeys, err := repository.scanKeys(requestContext, model.ThreadListCachePattern, cacheScanBatchSize)
	if err != nil {
		return err
	}

	if len(cacheKeys) == 0 {
		return nil
	}

	return repository.deleteKeys(requestContext, cacheKeys...)
}

func (repository *repository) InvalidateThreadDetailCache(requestContext context.Context, threadID int64) error {
	cacheKey := helper.BuildThreadDetailCacheKey(threadID)
	return repository.deleteKeys(requestContext, cacheKey)
}

func (repository *repository) deleteKeys(requestContext context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	err := repository.redisClient.Del(requestContext, keys...).Err()
	if err != nil {
		return fmt.Errorf("%w: delete cache keys: %v", model.ErrConsumeEvent, err)
	}

	return nil
}

func (repository *repository) scanKeys(
	requestContext context.Context,
	keyPattern string,
	batchSize int64,
) ([]string, error) {
	if batchSize <= 0 {
		return nil, fmt.Errorf("%w: redis batch size must be positive", model.ErrConsumeEvent)
	}

	var (
		nextCursor    uint64
		collectedKeys []string
	)

	for {
		cacheKeys, updatedCursor, err := repository.redisClient.Scan(
			requestContext,
			nextCursor,
			keyPattern,
			batchSize,
		).Result()
		if err != nil {
			return nil, fmt.Errorf("%w: scan cache keys: %v", model.ErrConsumeEvent, err)
		}

		collectedKeys = append(collectedKeys, cacheKeys...)
		nextCursor = updatedCursor
		if nextCursor == 0 {
			break
		}
	}

	return collectedKeys, nil
}

func wrapSQLError(action string, err error) error {
	if isInvalidPayloadSQLError(err) {
		return fmt.Errorf("%w: %s: %v", model.ErrInvalidEventPayload, action, err)
	}

	return fmt.Errorf("%w: %s: %v", model.ErrConsumeEvent, action, err)
}

func isInvalidPayloadSQLError(err error) bool {
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
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
