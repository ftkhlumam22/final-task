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

func GetThreadDetail(
	requestContext context.Context,
	databaseConnection *sql.DB,
	redisClient *redis.Client,
	threadID int64,
) (response.ThreadDetail, error) {
	cacheKey := repository.BuildThreadDetailCacheKey()
	cacheField := repository.BuildThreadDetailCacheField(threadID)

	cachedThreadDetail, isCacheHit, cacheReadError := repository.GetThreadDetailCache(requestContext, redisClient, cacheKey, cacheField)
	if cacheReadError != nil {
		log.Printf("[PRODUCER] Gagal baca cache detail thread. key=%s field=%s err=%v", cacheKey, cacheField, cacheReadError)
	}
	if isCacheHit {
		log.Printf("[PRODUCER] Detail thread diambil dari cache. key=%s field=%s", cacheKey, cacheField)
		return cachedThreadDetail, nil
	}

	threadDetail, getDetailError := repository.GetThreadDetail(requestContext, databaseConnection, threadID)
	if getDetailError != nil {
		return response.ThreadDetail{}, getDetailError
	}

	commentList, getCommentError := repository.GetThreadCommentTree(requestContext, databaseConnection, threadID)
	if getCommentError != nil {
		return response.ThreadDetail{}, getCommentError
	}
	threadDetail.CommentList = commentList

	cacheSetError := repository.SetThreadDetailCache(
		requestContext,
		redisClient,
		cacheKey,
		cacheField,
		threadDetail,
		time.Duration(model.DefaultThreadDetailCacheTTLSec)*time.Second,
	)
	if cacheSetError != nil {
		log.Printf("[PRODUCER] Gagal simpan cache detail thread. key=%s field=%s err=%v", cacheKey, cacheField, cacheSetError)
	} else {
		log.Printf("[PRODUCER] Detail thread berhasil disimpan ke cache. key=%s field=%s", cacheKey, cacheField)
	}

	return threadDetail, nil
}
