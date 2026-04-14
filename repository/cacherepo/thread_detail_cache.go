package cacherepo

import (
	"time"
)

func NewThreadDetailCacheRepository(redisRepository *RedisRepository) *ThreadDetailCacheRepository {
	return &ThreadDetailCacheRepository{redisRepository: redisRepository}
}

func (threadDetailCacheRepository *ThreadDetailCacheRepository) GetHashField(
	cacheKey string,
	cacheField string,
) (string, bool, error) {
	return threadDetailCacheRepository.redisRepository.HGet(cacheKey, cacheField)
}

func (threadDetailCacheRepository *ThreadDetailCacheRepository) SetHashFieldWithTTL(
	cacheKey string,
	cacheField string,
	cacheValue string,
	cacheTTL time.Duration,
) error {
	return threadDetailCacheRepository.redisRepository.HSetWithTTL(cacheKey, cacheField, cacheValue, cacheTTL)
}

func (threadDetailCacheRepository *ThreadDetailCacheRepository) DeleteHashFields(
	cacheKey string,
	cacheFields ...string,
) error {
	return threadDetailCacheRepository.redisRepository.HDeleteFields(cacheKey, cacheFields...)
}
