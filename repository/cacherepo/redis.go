package cacherepo

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisRepository(redisClient *redis.Client) *RedisRepository {
	return &RedisRepository{redisClient: redisClient}
}

func (redisRepository *RedisRepository) SetWithTTL(cacheKey string, cacheValue string, cacheTTL time.Duration) error {
	return redisRepository.redisClient.Set(context.Background(), cacheKey, cacheValue, cacheTTL).Err()
}

func (redisRepository *RedisRepository) Exists(cacheKey string) (bool, error) {
	existsCount, err := redisRepository.redisClient.Exists(context.Background(), cacheKey).Result()
	if err != nil {
		return false, err
	}

	return existsCount > 0, nil
}

func (redisRepository *RedisRepository) Delete(cacheKeys ...string) error {
	if len(cacheKeys) == 0 {
		return nil
	}

	return redisRepository.redisClient.Del(context.Background(), cacheKeys...).Err()
}

func (redisRepository *RedisRepository) HGet(cacheKey string, cacheField string) (string, bool, error) {
	cachedValue, err := redisRepository.redisClient.HGet(context.Background(), cacheKey, cacheField).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", false, nil
		}
		return "", false, err
	}

	return cachedValue, true, nil
}

func (redisRepository *RedisRepository) HSetWithTTL(cacheKey string, cacheField string, cacheValue string, cacheTTL time.Duration) error {
	err := redisRepository.redisClient.HSet(context.Background(), cacheKey, cacheField, cacheValue).Err()
	if err != nil {
		return err
	}

	return redisRepository.redisClient.Expire(context.Background(), cacheKey, cacheTTL).Err()
}

func (redisRepository *RedisRepository) HDeleteFields(cacheKey string, cacheFields ...string) error {
	if len(cacheFields) == 0 {
		return nil
	}

	return redisRepository.redisClient.HDel(context.Background(), cacheKey, cacheFields...).Err()
}

func (redisRepository *RedisRepository) ScanKeys(cursor uint64, pattern string, count int64) ([]string, uint64, error) {
	return redisRepository.redisClient.Scan(context.Background(), cursor, pattern, count).Result()
}
