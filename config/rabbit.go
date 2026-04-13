package config

import (
	"fmt"

	"kaktus-consumer/model"

	amqp "github.com/rabbitmq/amqp091-go"
)

func NewRabbitConsumer(rabbitConfig model.RabbitConfig) (*model.RabbitConsumer, error) {
	rabbitConnection, rabbitConnectionError := amqp.Dial(rabbitConfig.URL)
	if rabbitConnectionError != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", rabbitConnectionError)
	}

	rabbitChannel, rabbitChannelError := rabbitConnection.Channel()
	if rabbitChannelError != nil {
		_ = rabbitConnection.Close()
		return nil, fmt.Errorf("open channel: %w", rabbitChannelError)
	}

	if exchangeDeclareError := rabbitChannel.ExchangeDeclare(
		rabbitConfig.ExchangeName,
		rabbitConfig.ExchangeType,
		true,
		false,
		false,
		false,
		nil,
	); exchangeDeclareError != nil {
		_ = rabbitChannel.Close()
		_ = rabbitConnection.Close()
		return nil, fmt.Errorf("declare exchange: %w", exchangeDeclareError)
	}

	threadCreatedQueue, queueDeclareError := rabbitChannel.QueueDeclare(
		rabbitConfig.ThreadCreatedQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if queueDeclareError != nil {
		_ = rabbitChannel.Close()
		_ = rabbitConnection.Close()
		return nil, fmt.Errorf("declare queue: %w", queueDeclareError)
	}

	routingKeys := []string{
		rabbitConfig.ThreadCreatedRoutingKey,
		model.EventCommentCreated,
		model.EventThreadLiked,
	}
	for _, routingKey := range routingKeys {
		if queueBindError := rabbitChannel.QueueBind(
			threadCreatedQueue.Name,
			routingKey,
			rabbitConfig.ExchangeName,
			false,
			nil,
		); queueBindError != nil {
			_ = rabbitChannel.Close()
			_ = rabbitConnection.Close()
			return nil, fmt.Errorf("bind queue with routing key %s: %w", routingKey, queueBindError)
		}
	}

	if qosError := rabbitChannel.Qos(10, 0, false); qosError != nil {
		_ = rabbitChannel.Close()
		_ = rabbitConnection.Close()
		return nil, fmt.Errorf("set qos: %w", qosError)
	}

	return &model.RabbitConsumer{
		Connection: rabbitConnection,
		Channel:    rabbitChannel,
		ExchangeName: rabbitConfig.ExchangeName,
		QueueName:   threadCreatedQueue.Name,
		MaxRetry:    rabbitConfig.MaxRetry,
	}, nil
}
