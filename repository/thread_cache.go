package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"final-task/dto/response"
	"final-task/helper"
	"final-task/model"
	"github.com/redis/go-redis/v9"
)

func BuildThreadListCacheKey(limit int) string {
	return fmt.Sprintf("%s:%d", model.CacheThreadListHashKeyPrefix, limit)
}

func BuildThreadListCacheField(page int) string {
	return strconv.Itoa(page)
}

func GetThreadListCache(
	requestContext context.Context,
	redisClient *redis.Client,
	cacheKey string,
	cacheField string,
) (response.GetAllThread, bool, error) {
	cachedValue, isCacheHit, cacheError := helper.RedisHGet(requestContext, redisClient, cacheKey, cacheField)
	if cacheError != nil {
		return response.GetAllThread{}, false, cacheError
	}
	if !isCacheHit {
		return response.GetAllThread{}, false, nil
	}

	var cachedThreadList response.GetAllThread
	unmarshalError := json.Unmarshal([]byte(cachedValue), &cachedThreadList)
	if unmarshalError != nil {
		return response.GetAllThread{}, false, unmarshalError
	}

	return cachedThreadList, true, nil
}

func SetThreadListCache(
	requestContext context.Context,
	redisClient *redis.Client,
	cacheKey string,
	cacheField string,
	threadList response.GetAllThread,
	cacheTTL time.Duration,
) error {
	serializedValue, marshalError := json.Marshal(threadList)
	if marshalError != nil {
		return marshalError
	}

	return helper.RedisHSetWithTTL(requestContext, redisClient, cacheKey, cacheField, string(serializedValue), cacheTTL)
}

func InvalidateThreadListCache(requestContext context.Context, redisClient *redis.Client) error {
	var cursor uint64
	for {
		keys, nextCursor, scanError := helper.RedisScanKeys(requestContext, redisClient, cursor, model.CacheThreadListPattern, 100)
		if scanError != nil {
			return scanError
		}

		if len(keys) > 0 {
			deleteError := helper.RedisDelete(requestContext, redisClient, keys...)
			if deleteError != nil {
				return deleteError
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}
