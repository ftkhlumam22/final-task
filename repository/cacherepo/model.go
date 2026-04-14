package cacherepo

import "github.com/redis/go-redis/v9"

type RedisRepository struct {
	redisClient *redis.Client
}

type ThreadListCacheRepository struct {
	redisRepository *RedisRepository
}

type ThreadDetailCacheRepository struct {
	redisRepository *RedisRepository
}

type UserCacheRepository struct {
	redisRepository *RedisRepository
}
