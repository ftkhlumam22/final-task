package threadmodule

import (
	"fmt"
	"strconv"

	"final-task/model"
)

func buildThreadListCacheKey(limit int) string {
	return fmt.Sprintf("%s:%d", model.CacheThreadListHashKeyPrefix, limit)
}

func buildThreadListCacheField(page int) string {
	return strconv.Itoa(page)
}

func buildThreadDetailCacheField(threadID int64) string {
	return strconv.FormatInt(threadID, 10)
}

func (threadService *ThreadService) invalidateThreadListCache() error {
	var cursor uint64
	for {
		cacheKeys, nextCursor, err := threadService.threadListCacheRepository.ScanKeys(
			cursor,
			model.CacheThreadListPattern,
			100,
		)
		if err != nil {
			return err
		}

		if len(cacheKeys) > 0 {
			err := threadService.threadListCacheRepository.DeleteKeys(cacheKeys...)
			if err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

func (threadService *ThreadService) invalidateThreadDetailCache(threadID int64) error {
	cacheField := buildThreadDetailCacheField(threadID)
	return threadService.threadDetailCacheRepository.DeleteHashFields(model.CacheThreadDetailHashKey, cacheField)
}
