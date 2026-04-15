package threadrepo

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

func (threadCacheRepository *ThreadCacheRepository) GetHashField(
	cacheKey string,
	cacheField string,
) (string, bool, error) {
	cachedValue, err := threadCacheRepository.redisClient.HGet(context.Background(), cacheKey, cacheField).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", false, nil
		}
		return "", false, err
	}

	return cachedValue, true, nil
}

func (threadCacheRepository *ThreadCacheRepository) SetHashFieldWithTTL(
	cacheKey string,
	cacheField string,
	cacheValue string,
	cacheTTL time.Duration,
) error {
	err := threadCacheRepository.redisClient.HSet(context.Background(), cacheKey, cacheField, cacheValue).Err()
	if err != nil {
		return err
	}

	return threadCacheRepository.redisClient.Expire(context.Background(), cacheKey, cacheTTL).Err()
}

func (threadCacheRepository *ThreadCacheRepository) ScanKeys(
	cursor uint64,
	pattern string,
	count int64,
) ([]string, uint64, error) {
	return threadCacheRepository.redisClient.Scan(context.Background(), cursor, pattern, count).Result()
}

func (threadCacheRepository *ThreadCacheRepository) DeleteKeys(cacheKeys ...string) error {
	if len(cacheKeys) == 0 {
		return nil
	}

	return threadCacheRepository.redisClient.Del(context.Background(), cacheKeys...).Err()
}

func (threadCacheRepository *ThreadCacheRepository) DeleteHashFields(
	cacheKey string,
	cacheFields ...string,
) error {
	if len(cacheFields) == 0 {
		return nil
	}

	return threadCacheRepository.redisClient.HDel(context.Background(), cacheKey, cacheFields...).Err()
}
