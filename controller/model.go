package controller

import (
	"net/http"

	"final-task/dto/response"
	"final-task/model"
)

type AuthController struct {
	authService AuthService
}

type ThreadController struct {
	threadService ThreadService
}

type AuthService interface {
	RegisterUser(userName string, userEmail string, userPassword string) (model.User, error)
	LoginUser(userEmail string, userPassword string) (model.AuthResult, error)
	RefreshUserToken(refreshToken string) (model.AuthResult, error)
	LogoutUser(refreshToken string) error
}

type ThreadService interface {
	ListThread(page int, limit int) (response.GetAllThread, error)
	GetThreadDetail(threadID int64) (response.ThreadDetail, error)
	CreateThread(title string, description string, createdBy int64) error
	CreateComment(threadID int64, comment string, parentCommentID *int64, createdBy int64) error
	GetThreadLiked() ([]response.LikedThreadData, error)
	InsertLikeThread(threadID int64, likedBy int64) error
}

type AuthControllerHandler interface {
	RegisterHandler(responseWriter http.ResponseWriter, httpRequest *http.Request)
	LoginHandler(responseWriter http.ResponseWriter, httpRequest *http.Request)
	RefreshHandler(responseWriter http.ResponseWriter, httpRequest *http.Request)
	LogoutHandler(responseWriter http.ResponseWriter, httpRequest *http.Request)
}

type ThreadControllerHandler interface {
	ListThreadHandler(responseWriter http.ResponseWriter, httpRequest *http.Request)
	DetailThreadHandler(responseWriter http.ResponseWriter, httpRequest *http.Request)
	CreateThreadHandler(responseWriter http.ResponseWriter, httpRequest *http.Request)
	CreateCommentHandler(responseWriter http.ResponseWriter, httpRequest *http.Request)
	GetLikedThread(responseWriter http.ResponseWriter, httpRequest *http.Request)
	InsertLikeThreadHandler(responseWriter http.ResponseWriter, httpRequest *http.Request)
}
