package publisherrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"kaktus-consumer/model"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (repository *PublisherRepository) PublishRPCResponse(
	requestContext context.Context,
	replyToQueue string,
	correlationID string,
	responsePayload interface{},
) error {
	replyToQueue = strings.TrimSpace(replyToQueue)
	if replyToQueue == "" {
		return fmt.Errorf("%w: missing reply_to", model.ErrInvalidEventPayload)
	}

	responseBody, err := json.Marshal(responsePayload)
	if err != nil {
		return fmt.Errorf("%w: marshal rpc response: %v", model.ErrConsumeEvent, err)
	}

	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	channel, err := repository.getChannelLocked()
	if err != nil {
		return err
	}

	err = channel.PublishWithContext(
		requestContext,
		"",
		replyToQueue,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			Body:          responseBody,
		},
	)
	if err == nil {
		return nil
	}

	err = repository.resetChannelLocked()
	if err != nil {
		return fmt.Errorf("%w: publish rpc response: %v", model.ErrConsumeEvent, err)
	}

	channel, err = repository.getChannelLocked()
	if err != nil {
		return err
	}

	err = channel.PublishWithContext(
		requestContext,
		"",
		replyToQueue,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			Body:          responseBody,
		},
	)
	if err != nil {
		return fmt.Errorf("%w: publish rpc response: %v", model.ErrConsumeEvent, err)
	}

	return nil
}

func (repository *PublisherRepository) Close() error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if repository.channel == nil || repository.channel.IsClosed() {
		return nil
	}

	err := repository.channel.Close()
	if err != nil {
		return fmt.Errorf("%w: close publisher channel: %v", model.ErrConsumeEvent, err)
	}

	return nil
}

func (repository *PublisherRepository) getChannelLocked() (*amqp.Channel, error) {
	if repository.channel != nil && !repository.channel.IsClosed() {
		return repository.channel, nil
	}

	err := repository.resetChannelLocked()
	if err != nil {
		return nil, err
	}

	return repository.channel, nil
}

func (repository *PublisherRepository) resetChannelLocked() error {
	if repository.channel != nil && !repository.channel.IsClosed() {
		_ = repository.channel.Close()
	}

	channel, err := repository.connection.Channel()
	if err != nil {
		return fmt.Errorf("%w: open publisher channel: %v", model.ErrConsumeEvent, err)
	}

	repository.channel = channel
	return nil
}
