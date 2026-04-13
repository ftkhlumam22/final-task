package module

import (
	"strings"
	"unicode/utf8"

	"kaktus-consumer/model"
)

func ValidateThreadCreatedEvent(threadCreatedEvent model.ThreadCreatedEvent) error {
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
