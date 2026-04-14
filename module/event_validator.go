package module

import (
	"strings"
	"unicode/utf8"

	"kaktus-consumer/model"
)

func validateThreadCreatedEvent(threadCreatedEvent model.ThreadCreatedEvent) error {
	threadCreatedEvent.Title = strings.TrimSpace(threadCreatedEvent.Title)
	threadCreatedEvent.Description = strings.TrimSpace(threadCreatedEvent.Description)
	threadCreatedEvent.RequestID = strings.TrimSpace(threadCreatedEvent.RequestID)

	if !isThreadCreatedEventValid(threadCreatedEvent) {
		return model.ErrInvalidEventPayload
	}

	return nil
}

func isThreadCreatedEventValid(threadCreatedEvent model.ThreadCreatedEvent) bool {
	if threadCreatedEvent.RequestID == "" {
		return false
	}
	if threadCreatedEvent.Event != model.EventThreadCreated {
		return false
	}
	if threadCreatedEvent.CreatedBy <= 0 {
		return false
	}

	if strings.TrimSpace(threadCreatedEvent.Title) == "" {
		return false
	}
	if utf8.RuneCountInString(threadCreatedEvent.Title) > model.MaxThreadTitleLength {
		return false
	}

	if strings.TrimSpace(threadCreatedEvent.Description) == "" {
		return false
	}
	if utf8.RuneCountInString(threadCreatedEvent.Description) > model.MaxThreadDescriptionLength {
		return false
	}

	return true
}

func validateCommentCreatedEvent(commentCreatedEvent model.CommentCreatedEvent) error {
	commentCreatedEvent.Event = strings.TrimSpace(commentCreatedEvent.Event)
	commentCreatedEvent.RequestID = strings.TrimSpace(commentCreatedEvent.RequestID)
	commentCreatedEvent.Comment = strings.TrimSpace(commentCreatedEvent.Comment)

	if commentCreatedEvent.Event != model.EventCommentCreated ||
		commentCreatedEvent.RequestID == "" ||
		commentCreatedEvent.ThreadID <= 0 ||
		commentCreatedEvent.CreatedBy <= 0 ||
		commentCreatedEvent.Comment == "" ||
		utf8.RuneCountInString(commentCreatedEvent.Comment) > model.MaxCommentLength {
		return model.ErrInvalidEventPayload
	}

	if commentCreatedEvent.ParentCommentID != nil && *commentCreatedEvent.ParentCommentID <= 0 {
		return model.ErrInvalidEventPayload
	}

	return nil
}

func validateThreadLikedEvent(threadLikedEvent model.ThreadLikedEvent) error {
	threadLikedEvent.Event = strings.TrimSpace(threadLikedEvent.Event)
	threadLikedEvent.RequestID = strings.TrimSpace(threadLikedEvent.RequestID)

	if threadLikedEvent.Event != model.EventThreadLiked ||
		threadLikedEvent.RequestID == "" ||
		threadLikedEvent.ThreadID <= 0 ||
		threadLikedEvent.LikedBy <= 0 {
		return model.ErrInvalidEventPayload
	}

	return nil
}

func validateThreadGetLikedRequest(threadGetLikedRequest model.ThreadGetLikedRequest) error {
	threadGetLikedRequest.Event = strings.TrimSpace(threadGetLikedRequest.Event)
	threadGetLikedRequest.RequestID = strings.TrimSpace(threadGetLikedRequest.RequestID)

	if threadGetLikedRequest.Event != model.EventThreadGetLiked ||
		threadGetLikedRequest.RequestID == "" ||
		threadGetLikedRequest.Body == nil {
		return model.ErrInvalidEventPayload
	}

	return nil
}
