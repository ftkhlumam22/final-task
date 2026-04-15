package threadmodule

import (
	"time"

	"final-task/dto/response"
	"final-task/model"
)

type ThreadService struct {
	threadSQLRepository         ThreadSQLRepository
	threadListCacheRepository   ThreadListCacheRepository
	threadDetailCacheRepository ThreadDetailCacheRepository
	rabbitPublisher             *model.RabbitPublisher
	rabbitRPCClient             *model.RabbitRPCClient
}

type ThreadSQLRepository interface {
	GetThreadList(limit int, offset int) ([]response.ThreadData, int, error)
	GetThreadDetail(threadID int64) (response.ThreadDetail, error)
	GetThreadCommentRows(threadID int64) ([]model.ThreadCommentRow, error)
}

type ThreadListCacheRepository interface {
	GetHashField(cacheKey string, cacheField string) (string, bool, error)
	SetHashFieldWithTTL(cacheKey string, cacheField string, cacheValue string, cacheTTL time.Duration) error
	ScanKeys(cursor uint64, pattern string, count int64) ([]string, uint64, error)
	DeleteKeys(cacheKeys ...string) error
}

type ThreadDetailCacheRepository interface {
	GetHashField(cacheKey string, cacheField string) (string, bool, error)
	SetHashFieldWithTTL(cacheKey string, cacheField string, cacheValue string, cacheTTL time.Duration) error
	DeleteHashFields(cacheKey string, cacheFields ...string) error
}
