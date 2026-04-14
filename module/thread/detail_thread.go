package threadmodule

import (
	"encoding/json"
	"log"
	"time"

	"final-task/dto/response"
	"final-task/model"
)

func (threadService *ThreadService) GetThreadDetail(threadID int64) (response.ThreadDetail, error) {
	cacheField := buildThreadDetailCacheField(threadID)
	cachedThreadDetailValue, isCacheHit, err := threadService.threadDetailCacheRepository.GetHashField(
		model.CacheThreadDetailHashKey,
		cacheField,
	)
	if err != nil {
		log.Printf("[PRODUCER] Gagal baca cache detail thread. thread_id=%d err=%v", threadID, err)
	}
	if isCacheHit {
		var cachedThreadDetail response.ThreadDetail
		err := json.Unmarshal([]byte(cachedThreadDetailValue), &cachedThreadDetail)
		if err == nil {
			log.Printf("[PRODUCER] Detail thread diambil dari cache. thread_id=%d", threadID)
			return cachedThreadDetail, nil
		}
		log.Printf("[PRODUCER] Gagal parse cache detail thread. thread_id=%d err=%v", threadID, err)
	}

	threadDetail, err := threadService.threadReadRepository.GetThreadDetail(threadID)
	if err != nil {
		return response.ThreadDetail{}, err
	}

	commentRows, err := threadService.threadReadRepository.GetThreadCommentRows(threadID)
	if err != nil {
		return response.ThreadDetail{}, err
	}
	threadDetail.CommentList = buildCommentTree(commentRows)

	serializedValue, err := json.Marshal(threadDetail)
	if err != nil {
		log.Printf("[PRODUCER] Gagal serialize cache detail thread. thread_id=%d err=%v", threadID, err)
		return threadDetail, nil
	}

	err = threadService.threadDetailCacheRepository.SetHashFieldWithTTL(
		model.CacheThreadDetailHashKey,
		cacheField,
		string(serializedValue),
		time.Duration(model.DefaultThreadDetailCacheTTLSec)*time.Second,
	)
	if err != nil {
		log.Printf("[PRODUCER] Gagal simpan cache detail thread. thread_id=%d err=%v", threadID, err)
	} else {
		log.Printf("[PRODUCER] Detail thread berhasil disimpan ke cache. thread_id=%d", threadID)
	}

	return threadDetail, nil
}

func buildCommentTree(commentRows []model.ThreadCommentRow) []response.CommentData {
	commentByID := make(map[int64]model.ThreadCommentRow, len(commentRows))
	childByParentID := make(map[int64][]int64)
	topLevelCommentIDs := make([]int64, 0)

	for _, commentRow := range commentRows {
		commentByID[commentRow.ID] = commentRow
		if commentRow.ParentCommentID.Valid {
			parentID := commentRow.ParentCommentID.Int64
			childByParentID[parentID] = append(childByParentID[parentID], commentRow.ID)
			continue
		}

		topLevelCommentIDs = append(topLevelCommentIDs, commentRow.ID)
	}

	var buildReplyTree func(commentID int64) response.CommentData
	buildReplyTree = func(commentID int64) response.CommentData {
		commentRow := commentByID[commentID]
		childCommentIDs := childByParentID[commentID]
		replyList := make([]response.CommentData, 0, len(childCommentIDs))
		for _, childCommentID := range childCommentIDs {
			replyList = append(replyList, buildReplyTree(childCommentID))
		}

		commentData := response.CommentData{
			Comment:    commentRow.Comment,
			CommentBy:  commentRow.CommentBy,
			ReplyList:  replyList,
			TotalReply: commentRow.TotalReply,
		}
		if commentRow.CreatedAt.Valid {
			commentData.CreatedAt = commentRow.CreatedAt.Time
		}

		return commentData
	}

	commentList := make([]response.CommentData, 0, len(topLevelCommentIDs))
	for _, commentID := range topLevelCommentIDs {
		commentList = append(commentList, buildReplyTree(commentID))
	}

	return commentList
}
