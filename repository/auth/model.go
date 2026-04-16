package authrepo

import (
	"database/sql"

	"github.com/redis/go-redis/v9"
)

type Repositories struct {
	AuthSQL   AuthSQLRepository
	AuthCache AuthCacheRepository
}

type AuthSQLRepository struct {
	db *sql.DB
}

type AuthCacheRepository struct {
	redisClient *redis.Client
}

func NewRepositories(databaseConnection *sql.DB, redisClient *redis.Client) Repositories {
	return Repositories{
		AuthSQL:   AuthSQLRepository{db: databaseConnection},
		AuthCache: AuthCacheRepository{redisClient: redisClient},
	}
}
