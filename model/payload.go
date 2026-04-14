package model

import "time"

type RegisterUserRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type CreateThreadRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type LikeThreadRequest struct {
	ThreadID int64 `json:"thread_id"`
}

type CommentRequest struct {
	Comment         string `json:"comment"`
	ThreadID        int64  `json:"thread_id"`
	ParentCommentID *int64 `json:"parent_comment_id,omitempty"`
}

type GeneralSuccessResponse struct {
	Status     string    `json:"status"`
	Data       any       `json:"data"`
	Code       int       `json:"code"`
	AccessTime time.Time `json:"access_time"`
}

type GeneralErrorResponse struct {
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	Code       int       `json:"code"`
	AccessTime time.Time `json:"access_time"`
}

type LoginUserResponse struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RegisterUserResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type CommentDataResponse struct {
	Comment    string                `json:"comment"`
	CommentBy  string                `json:"comment_by"`
	CreatedAt  time.Time             `json:"created_at"`
	ReplyList  []CommentDataResponse `json:"reply_list"`
	TotalReply int                   `json:"total_reply"`
}

type ThreadDataResponse struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type GetAllThreadResponse struct {
	ListForum  []ThreadDataResponse `json:"list_forum"`
	TotalForum int                  `json:"total_forum"`
}

type ThreadDetailResponse struct {
	ID            int64                 `json:"id"`
	Title         string                `json:"title"`
	Description   string                `json:"description"`
	CreatedBy     string                `json:"created_by"`
	CreatedAt     time.Time             `json:"created_at"`
	TotalLikes    int                   `json:"total_likes"`
	TotalComments int                   `json:"total_comments"`
	CommentList   []CommentDataResponse `json:"comment_list"`
}

type LikedThreadDataResponse struct {
	Title   string `json:"title"`
	LikedBy string `json:"liked_by"`
}
