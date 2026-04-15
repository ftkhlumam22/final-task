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

type Repository interface {
	BeginTransaction(requestContext context.Context) (Repository, error)
	CommitTransaction() error
	RollbackTransaction() error

	InsertThread(requestContext context.Context, threadCreatedEvent model.ThreadCreatedEvent) (model.Thread, error)
	InsertComment(requestContext context.Context, commentCreatedEvent model.CommentCreatedEvent) error
	IncrementParentCommentReply(requestContext context.Context, parentCommentID int64, threadID int64) (int64, error)
	IncrementThreadTotalComment(requestContext context.Context, threadID int64) (int64, error)
	InsertThreadLike(requestContext context.Context, threadLikedEvent model.ThreadLikedEvent) (int64, error)
	IncrementThreadTotalLike(requestContext context.Context, threadID int64) (int64, error)
	GetLikedThreads(requestContext context.Context) ([]model.ThreadLikedListItem, error)
	InvalidateThreadListCache(requestContext context.Context) error
	InvalidateThreadDetailCache(requestContext context.Context, threadID int64) error
}

type repository struct {
	databaseConnection *sql.DB
	redisClient        *redis.Client
	transaction        *sql.Tx
}

func NewRepository(dependency Dependency) Repository {
	return &repository{
		databaseConnection: dependency.DatabaseConnection,
		redisClient:        dependency.RedisClient,
	}
}
