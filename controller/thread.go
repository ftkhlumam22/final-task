package controller

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

	"final-task/dto/request"
	"final-task/helper"
	"final-task/model"
	"final-task/module"
	"github.com/redis/go-redis/v9"
)

func CreateThreadHandler(rabbitPublisher *model.RabbitPublisher, redisClient *redis.Client) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		log.Printf("[PRODUCER] Request buat thread diterima.")

		var createThreadRequest request.CreateThread
		decodeError := helper.DecodeRequestBody(httpRequest, &createThreadRequest)
		if decodeError != nil {
			log.Printf("[PRODUCER] Payload thread tidak valid. err=%v", decodeError)
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

		createdByUserID, userIDError := helper.UserIDFromRequest(httpRequest)
		if userIDError != nil {
			log.Printf("[PRODUCER] User tidak terautentikasi saat create thread.")
			helper.WriteError(responseWriter, http.StatusUnauthorized, model.MessageUnauthorized)
			return
		}
		log.Printf("[PRODUCER] Validasi request thread lolos. user_id=%d", createdByUserID)

		createThreadError := module.CreateThread(
			httpRequest.Context(),
			rabbitPublisher,
			redisClient,
			createThreadRequest.Title,
			createThreadRequest.Description,
			createdByUserID,
		)
		if createThreadError != nil {
			log.Printf("[PRODUCER] Gagal mengirim event thread ke RabbitMQ. err=%v", createThreadError)
			helper.WriteMappedError(responseWriter, createThreadError)
			return
		}
		log.Printf("[PRODUCER] Event thread berhasil diantrikan.")

		helper.WriteSuccess(responseWriter, http.StatusAccepted, map[string]string{
			"message": model.MessageThreadCreateQueued,
		})
	}
}

func ListThreadHandler(databaseConnection *sql.DB, redisClient *redis.Client) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		page := parsePositiveIntQuery(httpRequest, "page", model.DefaultThreadListPage)
		limit := parsePositiveIntQuery(httpRequest, "limit", model.DefaultThreadListLimit)
		log.Printf("[PRODUCER] Request list thread diterima. page=%d limit=%d", page, limit)

		threadList, listError := module.ListThread(httpRequest.Context(), databaseConnection, redisClient, page, limit)
		if listError != nil {
			log.Printf("[PRODUCER] Gagal ambil list thread. err=%v", listError)
			helper.WriteMappedError(responseWriter, listError)
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
}

func DetailThreadHandler(databaseConnection *sql.DB, redisClient *redis.Client) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		threadID := parsePositiveInt64Query(httpRequest, "thread_id", 0)
		if threadID <= 0 {
			helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageThreadIDRequired)
			return
		}
		log.Printf("[PRODUCER] Request detail thread diterima. thread_id=%d", threadID)

		threadDetail, detailError := module.GetThreadDetail(httpRequest.Context(), databaseConnection, redisClient, threadID)
		if detailError != nil {
			log.Printf("[PRODUCER] Gagal ambil detail thread. thread_id=%d err=%v", threadID, detailError)
			helper.WriteMappedError(responseWriter, detailError)
			return
		}

		helper.WriteSuccess(responseWriter, http.StatusOK, threadDetail)
	}
}

func CreateCommentHandler(rabbitPublisher *model.RabbitPublisher, redisClient *redis.Client) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		log.Printf("[PRODUCER] Request buat comment diterima.")

		var createCommentRequest request.Comment
		decodeError := helper.DecodeRequestBody(httpRequest, &createCommentRequest)
		if decodeError != nil {
			log.Printf("[PRODUCER] Payload comment tidak valid. err=%v", decodeError)
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

		createdByUserID, userIDError := helper.UserIDFromRequest(httpRequest)
		if userIDError != nil {
			log.Printf("[PRODUCER] User tidak terautentikasi saat create comment.")
			helper.WriteError(responseWriter, http.StatusUnauthorized, model.MessageUnauthorized)
			return
		}

		createCommentError := module.CreateComment(
			httpRequest.Context(),
			rabbitPublisher,
			redisClient,
			createCommentRequest.ThreadID,
			createCommentRequest.Comment,
			createCommentRequest.ParentCommentID,
			createdByUserID,
		)
		if createCommentError != nil {
			log.Printf("[PRODUCER] Gagal mengirim event comment ke RabbitMQ. err=%v", createCommentError)
			helper.WriteMappedError(responseWriter, createCommentError)
			return
		}
		log.Printf("[PRODUCER] Event comment berhasil diantrikan.")

		helper.WriteSuccess(responseWriter, http.StatusAccepted, map[string]string{
			"message": model.MessageCommentCreateQueued,
		})
	}
}

func InsertLikeThreadHandler(rabbitPublisher *model.RabbitPublisher, redisClient *redis.Client) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		log.Printf("[PRODUCER] Request like thread diterima.")

		var likeThreadRequest request.LikeThread
		decodeError := helper.DecodeRequestBody(httpRequest, &likeThreadRequest)
		if decodeError != nil {
			log.Printf("[PRODUCER] Payload like tidak valid. err=%v", decodeError)
			helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageInvalidRequestBody)
			return
		}
		if likeThreadRequest.ThreadID <= 0 {
			log.Printf("[PRODUCER] Validasi like gagal: thread_id kosong.")
			helper.WriteError(responseWriter, http.StatusBadRequest, model.MessageLikePayloadNeeded)
			return
		}

		likedByUserID, userIDError := helper.UserIDFromRequest(httpRequest)
		if userIDError != nil {
			log.Printf("[PRODUCER] User tidak terautentikasi saat like thread.")
			helper.WriteError(responseWriter, http.StatusUnauthorized, model.MessageUnauthorized)
			return
		}

		insertLikeError := module.InsertLikeThread(
			httpRequest.Context(),
			rabbitPublisher,
			redisClient,
			likeThreadRequest.ThreadID,
			likedByUserID,
		)
		if insertLikeError != nil {
			log.Printf("[PRODUCER] Gagal mengirim event like ke RabbitMQ. err=%v", insertLikeError)
			helper.WriteMappedError(responseWriter, insertLikeError)
			return
		}
		log.Printf("[PRODUCER] Event like berhasil diantrikan.")

		helper.WriteSuccess(responseWriter, http.StatusAccepted, map[string]string{
			"message": model.MessageLikeCreateQueued,
		})
	}
}

func parsePositiveIntQuery(httpRequest *http.Request, queryKey string, fallback int) int {
	queryValue := strings.TrimSpace(httpRequest.URL.Query().Get(queryKey))
	if queryValue == "" {
		return fallback
	}

	parsedValue, parseError := strconv.Atoi(queryValue)
	if parseError != nil || parsedValue <= 0 {
		return fallback
	}

	return parsedValue
}

func parsePositiveInt64Query(httpRequest *http.Request, queryKey string, fallback int64) int64 {
	queryValue := strings.TrimSpace(httpRequest.URL.Query().Get(queryKey))
	if queryValue == "" {
		return fallback
	}

	parsedValue, parseError := strconv.ParseInt(queryValue, 10, 64)
	if parseError != nil || parsedValue <= 0 {
		return fallback
	}

	return parsedValue
}
