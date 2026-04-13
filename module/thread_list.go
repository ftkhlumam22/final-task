package module

import (
	"context"
	"database/sql"
	"log"
	"time"

	"final-task/dto/response"
	"final-task/model"
	"final-task/repository"
	"github.com/redis/go-redis/v9"
)

func ListThread(
	requestContext context.Context,
	databaseConnection *sql.DB,
	redisClient *redis.Client,
	page int,
	limit int,
) (response.GetAllThread, error) {
	cacheKey := repository.BuildThreadListCacheKey(limit)
	cacheField := repository.BuildThreadListCacheField(page)
	cachedThreadList, isCacheHit, cacheReadError := repository.GetThreadListCache(requestContext, redisClient, cacheKey, cacheField)
	if cacheReadError != nil {
		log.Printf("[PRODUCER] Gagal baca cache list thread. key=%s field=%s err=%v", cacheKey, cacheField, cacheReadError)
	}

	if isCacheHit {
		log.Printf("[PRODUCER] List thread diambil dari cache. key=%s field=%s", cacheKey, cacheField)
		return cachedThreadList, nil
	}

	offset := (page - 1) * limit
	threadList, totalForum, queryError := repository.GetThreadList(requestContext, databaseConnection, limit, offset)
	if queryError != nil {
		return response.GetAllThread{}, queryError
	}

	result := response.GetAllThread{
		ListForum:  threadList,
		TotalForum: totalForum,
	}

	cacheSetError := repository.SetThreadListCache(
		requestContext,
		redisClient,
		cacheKey,
		cacheField,
		result,
		time.Duration(model.DefaultThreadListCacheTTLSec)*time.Second,
	)
	if cacheSetError != nil {
		log.Printf("[PRODUCER] Gagal simpan cache list thread. key=%s field=%s err=%v", cacheKey, cacheField, cacheSetError)
	} else {
		log.Printf("[PRODUCER] List thread berhasil disimpan ke cache. key=%s field=%s", cacheKey, cacheField)
	}

	return result, nil
}
