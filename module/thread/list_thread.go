package threadmodule

import (
	"encoding/json"
	"log"
	"time"

	"final-task/dto/response"
	"final-task/model"
)

func (threadService *ThreadService) ListThread(page int, limit int) (response.GetAllThread, error) {
	cacheKey := buildThreadListCacheKey(limit)
	cacheField := buildThreadListCacheField(page)

	cachedThreadListValue, isCacheHit, err := threadService.threadListCacheRepository.GetHashField(cacheKey, cacheField)
	if err != nil {
		log.Printf("[PRODUCER] Gagal baca cache list thread. page=%d limit=%d err=%v", page, limit, err)
	}
	if isCacheHit {
		var cachedThreadList response.GetAllThread
		err := json.Unmarshal([]byte(cachedThreadListValue), &cachedThreadList)
		if err == nil {
			log.Printf("[PRODUCER] List thread diambil dari cache. page=%d limit=%d", page, limit)
			return cachedThreadList, nil
		}
		log.Printf("[PRODUCER] Gagal parse cache list thread. page=%d limit=%d err=%v", page, limit, err)
	}

	offset := (page - 1) * limit
	threadList, totalForum, err := threadService.threadReadRepository.GetThreadList(limit, offset)
	if err != nil {
		return response.GetAllThread{}, err
	}

	result := response.GetAllThread{
		ListForum:  threadList,
		TotalForum: totalForum,
	}

	serializedValue, err := json.Marshal(result)
	if err != nil {
		log.Printf("[PRODUCER] Gagal serialize cache list thread. page=%d limit=%d err=%v", page, limit, err)
		return result, nil
	}

	err = threadService.threadListCacheRepository.SetHashFieldWithTTL(
		cacheKey,
		cacheField,
		string(serializedValue),
		time.Duration(model.DefaultThreadListCacheTTLSec)*time.Second,
	)
	if err != nil {
		log.Printf("[PRODUCER] Gagal simpan cache list thread. page=%d limit=%d err=%v", page, limit, err)
	} else {
		log.Printf("[PRODUCER] List thread berhasil disimpan ke cache. page=%d limit=%d", page, limit)
	}

	return result, nil
}
