package cacherepo

import "github.com/redis/go-redis/v9"

type CacheRepository struct {
	redisClient *redis.Client
}

func NewCacheRepository(redisClient *redis.Client) *CacheRepository {
	return &CacheRepository{
		redisClient: redisClient,
	}
}
