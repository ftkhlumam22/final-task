package repository

import (
	"context"
	"fmt"
	"log"

	"kaktus-consumer/helper"
	"kaktus-consumer/model"

	"github.com/redis/go-redis/v9"
)

func InvalidateThreadListCache(requestContext context.Context, redisClient *redis.Client) error {
	cacheKeys, scanError := helper.RedisScanKeys(requestContext, redisClient, model.ThreadListCachePattern, 100)
	if scanError != nil {
		log.Printf("[cache][thread-list] gagal scan cache list. pattern=%s error=%v", model.ThreadListCachePattern, scanError)
		return fmt.Errorf("%w: invalidate thread list cache: %v", model.ErrConsumeEvent, scanError)
	}

	if len(cacheKeys) > 0 {
		if deleteError := helper.RedisDelete(requestContext, redisClient, cacheKeys...); deleteError != nil {
			log.Printf("[cache][thread-list] gagal hapus cache list. keys=%v error=%v", cacheKeys, deleteError)
			return fmt.Errorf("%w: invalidate thread list cache: %v", model.ErrConsumeEvent, deleteError)
		}
	}

	return nil
}

func InvalidateThreadDetailCache(
	requestContext context.Context,
	redisClient *redis.Client,
	threadID int64,
) error {
	detailCacheKey := helper.BuildThreadDetailCacheKey(threadID)
	if deleteError := helper.RedisDelete(requestContext, redisClient, detailCacheKey); deleteError != nil {
		log.Printf("[cache][thread-detail] gagal hapus cache detail. key=%s error=%v", detailCacheKey, deleteError)
		return fmt.Errorf("%w: invalidate thread detail cache: %v", model.ErrConsumeEvent, deleteError)
	}

	return nil
}
