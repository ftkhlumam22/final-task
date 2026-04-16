package controller

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"final-task/dto/request"
	"final-task/helper"
	"final-task/middleware"
	"final-task/model"
)

func NewThreadController(threadService ThreadService) ThreadController {
	return ThreadController{threadService: threadService}
}

func (threadController ThreadController) GetLikedThread(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	log.Printf("[PRODUCER] Request list liked thread diterima.")

	likedThreadList, err := threadController.threadService.GetThreadLiked()
	if err != nil {
		log.Printf("[PRODUCER] Gagal ambil list liked thread. err=%v", err)
		helper.WriteMappedError(responseWriter, err)
		return
	}

	helper.WriteSuccess(responseWriter, http.StatusOK, map[string]any{
		"list_liked_thread":  likedThreadList,
		"total_liked_thread": len(likedThreadList),
	})
}

func (threadController ThreadController) CreateThreadHandler(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	log.Printf("[PRODUCER] Request buat thread diterima.")

	var createThreadRequest request.CreateThread
	err := helper.DecodeRequestBody(httpRequest, &createThreadRequest)
	if err != nil {
		log.Printf("[PRODUCER] Payload thread tidak valid. err=%v", err)
		helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageInvalidRequestBody)
		return
	}

	createThreadRequest.Title = strings.TrimSpace(createThreadRequest.Title)
	createThreadRequest.Description = strings.TrimSpace(createThreadRequest.Description)
	if createThreadRequest.Title == "" || createThreadRequest.Description == "" {
		log.Printf("[PRODUCER] Validasi thread gagal: title/description kosong.")
		helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageThreadPayloadNeeded)
		return
	}

	createdByUserID, err := middleware.UserIDFromRequest(httpRequest)
	if err != nil {
		log.Printf("[PRODUCER] User tidak terautentikasi saat create thread.")
		helper.WriteError(responseWriter, http.StatusUnauthorized, model.MessageUnauthorized)
		return
	}
	log.Printf("[PRODUCER] Validasi request thread lolos. user_id=%d", createdByUserID)

	err = threadController.threadService.CreateThread(
		createThreadRequest.Title,
		createThreadRequest.Description,
		createdByUserID,
	)
	if err != nil {
		log.Printf("[PRODUCER] Gagal mengirim event thread ke RabbitMQ. err=%v", err)
		helper.WriteMappedError(responseWriter, err)
		return
	}
	log.Printf("[PRODUCER] Event thread berhasil diantrikan.")

	helper.WriteSuccess(responseWriter, http.StatusAccepted, map[string]string{
		"message": model.MessageThreadCreateQueued,
	})
}

func (threadController ThreadController) ListThreadHandler(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	page := parsePositiveIntQuery(httpRequest, "page", model.DefaultThreadListPage)
	limit := parsePositiveIntQuery(httpRequest, "limit", model.DefaultThreadListLimit)
	log.Printf("[PRODUCER] Request list thread diterima. page=%d limit=%d", page, limit)

	threadList, err := threadController.threadService.ListThread(page, limit)
	if err != nil {
		log.Printf("[PRODUCER] Gagal ambil list thread. err=%v", err)
		helper.WriteMappedError(responseWriter, err)
		return
	}
	log.Printf("[PRODUCER] List thread berhasil diambil. total=%d", threadList.TotalForum)

	responsePayload := map[string]any{
		"list_forum":  threadList.ListForum,
		"total_forum": threadList.TotalForum,
		"page":        page,
		"limit":       limit,
	}
	helper.WriteSuccess(responseWriter, http.StatusOK, responsePayload)
}

func (threadController ThreadController) DetailThreadHandler(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	threadID := parsePositiveInt64Query(httpRequest, "thread_id", 0)
	if threadID <= 0 {
		helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageThreadIDRequired)
		return
	}
	log.Printf("[PRODUCER] Request detail thread diterima. thread_id=%d", threadID)

	threadDetail, err := threadController.threadService.GetThreadDetail(threadID)
	if err != nil {
		log.Printf("[PRODUCER] Gagal ambil detail thread. thread_id=%d err=%v", threadID, err)
		helper.WriteMappedError(responseWriter, err)
		return
	}

	helper.WriteSuccess(responseWriter, http.StatusOK, threadDetail)
}

func (threadController ThreadController) CreateCommentHandler(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	log.Printf("[PRODUCER] Request buat comment diterima.")

	var createCommentRequest request.Comment
	err := helper.DecodeRequestBody(httpRequest, &createCommentRequest)
	if err != nil {
		log.Printf("[PRODUCER] Payload comment tidak valid. err=%v", err)
		helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageInvalidRequestBody)
		return
	}

	createCommentRequest.Comment = strings.TrimSpace(createCommentRequest.Comment)
	if createCommentRequest.ThreadID <= 0 || createCommentRequest.Comment == "" {
		log.Printf("[PRODUCER] Validasi comment gagal: thread_id/comment kosong.")
		helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageCommentPayloadNeeded)
		return
	}
	if createCommentRequest.ParentCommentID != nil && *createCommentRequest.ParentCommentID <= 0 {
		log.Printf("[PRODUCER] Validasi comment gagal: parent_comment_id tidak valid.")
		helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageCommentPayloadNeeded)
		return
	}

	createdByUserID, err := middleware.UserIDFromRequest(httpRequest)
	if err != nil {
		log.Printf("[PRODUCER] User tidak terautentikasi saat create comment.")
		helper.WriteError(responseWriter, http.StatusUnauthorized, model.MessageUnauthorized)
		return
	}

	err = threadController.threadService.CreateComment(
		createCommentRequest.ThreadID,
		createCommentRequest.Comment,
		createCommentRequest.ParentCommentID,
		createdByUserID,
	)
	if err != nil {
		log.Printf("[PRODUCER] Gagal mengirim event comment ke RabbitMQ. err=%v", err)
		helper.WriteMappedError(responseWriter, err)
		return
	}
	log.Printf("[PRODUCER] Event comment berhasil diantrikan.")

	helper.WriteSuccess(responseWriter, http.StatusAccepted, map[string]string{
		"message": model.MessageCommentCreateQueued,
	})
}

func (threadController ThreadController) InsertLikeThreadHandler(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	log.Printf("[PRODUCER] Request like thread diterima.")

	var likeThreadRequest request.LikeThread
	err := helper.DecodeRequestBody(httpRequest, &likeThreadRequest)
	if err != nil {
		log.Printf("[PRODUCER] Payload like tidak valid. err=%v", err)
		helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageInvalidRequestBody)
		return
	}
	if likeThreadRequest.ThreadID <= 0 {
		log.Printf("[PRODUCER] Validasi like gagal: thread_id kosong.")
		helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageLikePayloadNeeded)
		return
	}

	likedByUserID, err := middleware.UserIDFromRequest(httpRequest)
	if err != nil {
		log.Printf("[PRODUCER] User tidak terautentikasi saat like thread.")
		helper.WriteError(responseWriter, http.StatusUnauthorized, model.MessageUnauthorized)
		return
	}

	err = threadController.threadService.InsertLikeThread(
		likeThreadRequest.ThreadID,
		likedByUserID,
	)
	if err != nil {
		log.Printf("[PRODUCER] Gagal mengirim event like ke RabbitMQ. err=%v", err)
		helper.WriteMappedError(responseWriter, err)
		return
	}
	log.Printf("[PRODUCER] Event like berhasil diantrikan.")

	helper.WriteSuccess(responseWriter, http.StatusAccepted, map[string]string{
		"message": model.MessageLikeCreateQueued,
	})
}

func parsePositiveIntQuery(httpRequest *http.Request, queryKey string, fallback int) int {
	queryValue := strings.TrimSpace(httpRequest.URL.Query().Get(queryKey))
	if queryValue == "" {
		return fallback
	}

	parsedValue, err := strconv.Atoi(queryValue)
	if err != nil || parsedValue <= 0 {
		return fallback
	}

	return parsedValue
}

func parsePositiveInt64Query(httpRequest *http.Request, queryKey string, fallback int64) int64 {
	queryValue := strings.TrimSpace(httpRequest.URL.Query().Get(queryKey))
	if queryValue == "" {
		return fallback
	}

	parsedValue, err := strconv.ParseInt(queryValue, 10, 64)
	if err != nil || parsedValue <= 0 {
		return fallback
	}

	return parsedValue
}
