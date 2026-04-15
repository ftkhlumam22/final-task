package authrepo

import (
	"context"
	"time"
)

func (authCacheRepository *AuthCacheRepository) Set(
	cacheKey string,
	cacheValue string,
	cacheTTL time.Duration,
) error {
	return authCacheRepository.redisClient.Set(context.Background(), cacheKey, cacheValue, cacheTTL).Err()
}

func (authCacheRepository *AuthCacheRepository) Exists(cacheKey string) (bool, error) {
	existsCount, err := authCacheRepository.redisClient.Exists(context.Background(), cacheKey).Result()
	if err != nil {
		return false, err
	}

	return existsCount > 0, nil
}

func (authCacheRepository *AuthCacheRepository) Delete(cacheKey string) error {
	return authCacheRepository.redisClient.Del(context.Background(), cacheKey).Err()
}
