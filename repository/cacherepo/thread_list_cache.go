package cacherepo

import (
	"time"
)

func NewThreadListCacheRepository(redisRepository *RedisRepository) *ThreadListCacheRepository {
	return &ThreadListCacheRepository{redisRepository: redisRepository}
}

func (threadListCacheRepository *ThreadListCacheRepository) GetHashField(
	cacheKey string,
	cacheField string,
) (string, bool, error) {
	return threadListCacheRepository.redisRepository.HGet(cacheKey, cacheField)
}

func (threadListCacheRepository *ThreadListCacheRepository) SetHashFieldWithTTL(
	cacheKey string,
	cacheField string,
	cacheValue string,
	cacheTTL time.Duration,
) error {
	return threadListCacheRepository.redisRepository.HSetWithTTL(cacheKey, cacheField, cacheValue, cacheTTL)
}

func (threadListCacheRepository *ThreadListCacheRepository) ScanKeys(
	cursor uint64,
	pattern string,
	count int64,
) ([]string, uint64, error) {
	return threadListCacheRepository.redisRepository.ScanKeys(cursor, pattern, count)
}

func (threadListCacheRepository *ThreadListCacheRepository) DeleteKeys(cacheKeys ...string) error {
	return threadListCacheRepository.redisRepository.Delete(cacheKeys...)
}
