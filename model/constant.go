package model

const (
	StatusSuccess = "success"
	StatusError   = "error"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

const (
	EventThreadCreated  = "thread.created"
	EventCommentCreated = "comment.created"
	EventThreadLiked    = "thread.liked"
)

const (
	CacheThreadListHashKeyPrefix = "thread:list:limit"
	CacheThreadListPattern       = "thread:list:limit:*"
	CacheThreadDetailHashKey     = "detail"
)

const (
	MessageInvalidRequestBody      = "invalid request body"
	MessageNameEmailPasswordNeeded = "name, email, and password are required"
	MessageEmailPasswordNeeded     = "email and password are required"
	MessageRefreshTokenNeeded      = "refresh_token is required"
	MessageThreadPayloadNeeded     = "title and description are required"
	MessageCommentPayloadNeeded    = "thread_id and comment are required"
	MessageLikePayloadNeeded       = "thread_id is required"
	MessageThreadIDRequired        = "thread_id is required"
	MessageLogoutSuccess           = "logout success"
	MessageThreadCreateQueued      = "thread creation queued"
	MessageCommentCreateQueued     = "comment creation queued"
	MessageLikeCreateQueued        = "like insertion queued"
	MessageInternalServerError     = "internal server error"
	MessageMethodNotAllowed        = "method not allowed"
	MessageJWTSecretRequired       = "JWT_SECRET_KEY is required"
	MessageUnauthorized            = "unauthorized"
	MessageThreadNotFound          = "thread not found"
)

const (
	MessageEmailAlreadyExists = "email already exists"
	MessageInvalidCredential  = "invalid credential"
	MessageInvalidRefresh     = "invalid refresh token"
	MessageInvalidToken       = "invalid token"
	MessageInvalidTokenType   = "invalid token type"
	MessageUnexpectedSigning  = "unexpected signing method"
	MessageFailedPublishEvent = "failed to publish event"
)
