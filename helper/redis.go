package helper

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

func RedisSetWithTTL(
	requestContext context.Context,
	redisClient *redis.Client,
	cacheKey string,
	cacheValue string,
	cacheTTL time.Duration,
) error {
	return redisClient.Set(requestContext, cacheKey, cacheValue, cacheTTL).Err()
}

func RedisExists(
	requestContext context.Context,
	redisClient *redis.Client,
	cacheKey string,
) (bool, error) {
	existsCount, existsError := redisClient.Exists(requestContext, cacheKey).Result()
	if existsError != nil {
		return false, existsError
	}

	return existsCount > 0, nil
}

func RedisDelete(
	requestContext context.Context,
	redisClient *redis.Client,
	cacheKeys ...string,
) error {
	if len(cacheKeys) == 0 {
		return nil
	}

	return redisClient.Del(requestContext, cacheKeys...).Err()
}

func RedisHGet(
	requestContext context.Context,
	redisClient *redis.Client,
	cacheKey string,
	cacheField string,
) (string, bool, error) {
	cachedValue, cacheReadError := redisClient.HGet(requestContext, cacheKey, cacheField).Result()
	if cacheReadError != nil {
		if errors.Is(cacheReadError, redis.Nil) {
			return "", false, nil
		}
		return "", false, cacheReadError
	}

	return cachedValue, true, nil
}

func RedisHSetWithTTL(
	requestContext context.Context,
	redisClient *redis.Client,
	cacheKey string,
	cacheField string,
	cacheValue string,
	cacheTTL time.Duration,
) error {
	hashSetError := redisClient.HSet(requestContext, cacheKey, cacheField, cacheValue).Err()
	if hashSetError != nil {
		return hashSetError
	}

	return redisClient.Expire(requestContext, cacheKey, cacheTTL).Err()
}

func RedisHDeleteFields(
	requestContext context.Context,
	redisClient *redis.Client,
	cacheKey string,
	cacheFields ...string,
) error {
	if len(cacheFields) == 0 {
		return nil
	}

	return redisClient.HDel(requestContext, cacheKey, cacheFields...).Err()
}

func RedisScanKeys(
	requestContext context.Context,
	redisClient *redis.Client,
	cursor uint64,
	pattern string,
	count int64,
) ([]string, uint64, error) {
	return redisClient.Scan(requestContext, cursor, pattern, count).Result()
}
