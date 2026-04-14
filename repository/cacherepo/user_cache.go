package cacherepo

import "time"

func NewUserCacheRepository(redisRepository *RedisRepository) *UserCacheRepository {
	return &UserCacheRepository{redisRepository: redisRepository}
}

func (userCacheRepository *UserCacheRepository) Set(
	cacheKey string,
	cacheValue string,
	cacheTTL time.Duration,
) error {
	return userCacheRepository.redisRepository.SetWithTTL(cacheKey, cacheValue, cacheTTL)
}

func (userCacheRepository *UserCacheRepository) Exists(cacheKey string) (bool, error) {
	return userCacheRepository.redisRepository.Exists(cacheKey)
}

func (userCacheRepository *UserCacheRepository) Delete(cacheKey string) error {
	return userCacheRepository.redisRepository.Delete(cacheKey)
}
