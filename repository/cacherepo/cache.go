package cacherepo

import (
	"context"
	"fmt"

	"kaktus-consumer/model"
)

func (repository *CacheRepository) Delete(
	requestContext context.Context,
	keys ...string,
) error {
	if len(keys) == 0 {
		return nil
	}

	err := repository.redisClient.Del(requestContext, keys...).Err()
	if err != nil {
		return fmt.Errorf("%w: delete cache keys: %v", model.ErrConsumeEvent, err)
	}

	return nil
}

func (repository *CacheRepository) ScanKeys(
	requestContext context.Context,
	keyPattern string,
	batchSize int64,
) ([]string, error) {
	if batchSize <= 0 {
		return nil, fmt.Errorf("%w: redis batch size must be positive", model.ErrConsumeEvent)
	}

	var (
		nextCursor    uint64
		collectedKeys []string
	)

	for {
		cacheKeys, updatedCursor, err := repository.redisClient.Scan(
			requestContext,
			nextCursor,
			keyPattern,
			batchSize,
		).Result()
		if err != nil {
			return nil, fmt.Errorf("%w: scan cache keys: %v", model.ErrConsumeEvent, err)
		}

		collectedKeys = append(collectedKeys, cacheKeys...)
		nextCursor = updatedCursor
		if nextCursor == 0 {
			break
		}
	}

	return collectedKeys, nil
}
