package module

import (
	"encoding/json"
	"strings"

	"kaktus-consumer/model"
)

func (service *eventService) ParseEventEnvelope(messageBody []byte) (model.EventEnvelope, error) {
	var eventEnvelope model.EventEnvelope
	err := json.Unmarshal(messageBody, &eventEnvelope)
	if err != nil {
		return model.EventEnvelope{}, model.ErrInvalidEventPayload
	}

	eventEnvelope.Event = strings.TrimSpace(eventEnvelope.Event)
	eventEnvelope.RequestID = strings.TrimSpace(eventEnvelope.RequestID)
	if eventEnvelope.Event == "" || eventEnvelope.RequestID == "" {
		return model.EventEnvelope{}, model.ErrInvalidEventPayload
	}

	return eventEnvelope, nil
}

func parseThreadCreatedEvent(messageBody []byte) (model.ThreadCreatedEvent, error) {
	var threadCreatedEvent model.ThreadCreatedEvent
	err := json.Unmarshal(messageBody, &threadCreatedEvent)
	if err != nil {
		return model.ThreadCreatedEvent{}, model.ErrInvalidEventPayload
	}

	return threadCreatedEvent, nil
}

func parseCommentCreatedEvent(messageBody []byte) (model.CommentCreatedEvent, error) {
	var commentCreatedEvent model.CommentCreatedEvent
	err := json.Unmarshal(messageBody, &commentCreatedEvent)
	if err != nil {
		return model.CommentCreatedEvent{}, model.ErrInvalidEventPayload
	}

	return commentCreatedEvent, nil
}

func parseThreadLikedEvent(messageBody []byte) (model.ThreadLikedEvent, error) {
	var threadLikedEvent model.ThreadLikedEvent
	err := json.Unmarshal(messageBody, &threadLikedEvent)
	if err != nil {
		return model.ThreadLikedEvent{}, model.ErrInvalidEventPayload
	}

	return threadLikedEvent, nil
}

func parseThreadGetLikedRequest(messageBody []byte) (model.ThreadGetLikedRequest, error) {
	var threadGetLikedRequest model.ThreadGetLikedRequest
	err := json.Unmarshal(messageBody, &threadGetLikedRequest)
	if err != nil {
		return model.ThreadGetLikedRequest{}, model.ErrInvalidEventPayload
	}

	return threadGetLikedRequest, nil
}
