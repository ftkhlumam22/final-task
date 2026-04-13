package repository

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"final-task/dto/response"
	"final-task/helper"
	"final-task/model"
	"github.com/redis/go-redis/v9"
)

func BuildThreadDetailCacheKey() string {
	return model.CacheThreadDetailHashKey
}

func BuildThreadDetailCacheField(threadID int64) string {
	return strconv.FormatInt(threadID, 10)
}

func GetThreadDetailCache(
	requestContext context.Context,
	redisClient *redis.Client,
	cacheKey string,
	cacheField string,
) (response.ThreadDetail, bool, error) {
	cachedValue, isCacheHit, cacheError := helper.RedisHGet(requestContext, redisClient, cacheKey, cacheField)
	if cacheError != nil {
		return response.ThreadDetail{}, false, cacheError
	}
	if !isCacheHit {
		return response.ThreadDetail{}, false, nil
	}

	var cachedThreadDetail response.ThreadDetail
	unmarshalError := json.Unmarshal([]byte(cachedValue), &cachedThreadDetail)
	if unmarshalError != nil {
		return response.ThreadDetail{}, false, unmarshalError
	}

	return cachedThreadDetail, true, nil
}

func SetThreadDetailCache(
	requestContext context.Context,
	redisClient *redis.Client,
	cacheKey string,
	cacheField string,
	threadDetail response.ThreadDetail,
	cacheTTL time.Duration,
) error {
	serializedValue, marshalError := json.Marshal(threadDetail)
	if marshalError != nil {
		return marshalError
	}

	return helper.RedisHSetWithTTL(requestContext, redisClient, cacheKey, cacheField, string(serializedValue), cacheTTL)
}

func InvalidateThreadDetailCache(
	requestContext context.Context,
	redisClient *redis.Client,
	threadID int64,
) error {
	cacheKey := BuildThreadDetailCacheKey()
	cacheField := BuildThreadDetailCacheField(threadID)
	return helper.RedisHDeleteFields(requestContext, redisClient, cacheKey, cacheField)
}
