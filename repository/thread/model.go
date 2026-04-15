package threadrepo

import (
	"database/sql"

	"github.com/redis/go-redis/v9"
)

type Repositories struct {
	ThreadSQL   *ThreadSQLRepository
	ThreadCache *ThreadCacheRepository
}

type ThreadSQLRepository struct {
	db *sql.DB
}

type ThreadCacheRepository struct {
	redisClient *redis.Client
}

func NewRepositories(databaseConnection *sql.DB, redisClient *redis.Client) *Repositories {
	return &Repositories{
		ThreadSQL:   &ThreadSQLRepository{db: databaseConnection},
		ThreadCache: &ThreadCacheRepository{redisClient: redisClient},
	}
}
