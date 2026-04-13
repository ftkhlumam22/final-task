package helper

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func RedisSetWithTTL(
	requestContext context.Context,
	redisClient *redis.Client,
	key string,
	value interface{},
	ttl time.Duration,
) error {
	if setError := redisClient.Set(requestContext, key, value, ttl).Err(); setError != nil {
		return setError
	}
	return nil
}

func RedisExists(
	requestContext context.Context,
	redisClient *redis.Client,
	key string,
) (bool, error) {
	totalKey, existsError := redisClient.Exists(requestContext, key).Result()
	if existsError != nil {
		return false, existsError
	}
	return totalKey > 0, nil
}

func RedisDelete(
	requestContext context.Context,
	redisClient *redis.Client,
	keys ...string,
) error {
	if len(keys) == 0 {
		return nil
	}
	if deleteError := redisClient.Del(requestContext, keys...).Err(); deleteError != nil {
		return deleteError
	}
	return nil
}

func RedisScanKeys(
	requestContext context.Context,
	redisClient *redis.Client,
	keyPattern string,
	batchSize int64,
) ([]string, error) {
	if batchSize <= 0 {
		return nil, fmt.Errorf("batch size must be positive")
	}

	var (
		nextCursor uint64
		collectedKeys []string
	)

	for {
		cacheKeys, updatedCursor, scanError := redisClient.Scan(
			requestContext,
			nextCursor,
			keyPattern,
			batchSize,
		).Result()
		if scanError != nil {
			return nil, scanError
		}

		collectedKeys = append(collectedKeys, cacheKeys...)
		nextCursor = updatedCursor
		if nextCursor == 0 {
			break
		}
	}

	return collectedKeys, nil
}
