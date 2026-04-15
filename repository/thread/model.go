package thread

import (
	"context"
	"database/sql"

	"kaktus-consumer/model"

	"github.com/redis/go-redis/v9"
)

type Dependency struct {
	DatabaseConnection *sql.DB
	RedisClient        *redis.Client
}

type TransactionRepository interface {
	InsertComment(requestContext context.Context, commentCreatedEvent model.CommentCreatedEvent) error
	IncrementParentCommentReply(requestContext context.Context, parentCommentID int64, threadID int64) (int64, error)
	IncrementThreadTotalComment(requestContext context.Context, threadID int64) (int64, error)
	InsertThreadLike(requestContext context.Context, threadLikedEvent model.ThreadLikedEvent) (int64, error)
	IncrementThreadTotalLike(requestContext context.Context, threadID int64) (int64, error)
}

type Repository interface {
	InsertThread(requestContext context.Context, threadCreatedEvent model.ThreadCreatedEvent) (model.Thread, error)
	WithTransaction(requestContext context.Context, operation func(transactionRepository TransactionRepository) error) error
	GetLikedThreads(requestContext context.Context) ([]model.ThreadLikedListItem, error)
	InvalidateThreadListCache(requestContext context.Context) error
	InvalidateThreadDetailCache(requestContext context.Context, threadID int64) error
}

type repository struct {
	databaseConnection *sql.DB
	redisClient        *redis.Client
}

type sqlTransactionRepository struct {
	transaction *sql.Tx
}

func NewRepository(dependency Dependency) Repository {
	return &repository{
		databaseConnection: dependency.DatabaseConnection,
		redisClient:        dependency.RedisClient,
	}
}
