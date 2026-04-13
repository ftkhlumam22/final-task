package config

import (
	"fmt"

	"final-task/model"
	amqp "github.com/rabbitmq/amqp091-go"
)

func NewRabbitPublisher(rabbitConfig model.RabbitConfig) (*model.RabbitPublisher, error) {
	connection, connectionError := amqp.Dial(rabbitConfig.URL)
	if connectionError != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", connectionError)
	}

	channel, channelError := connection.Channel()
	if channelError != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("open rabbitmq channel: %w", channelError)
	}

	exchangeError := channel.ExchangeDeclare(
		rabbitConfig.ExchangeName,
		rabbitConfig.ExchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	if exchangeError != nil {
		_ = channel.Close()
		_ = connection.Close()
		return nil, fmt.Errorf("declare exchange: %w", exchangeError)
	}

	return &model.RabbitPublisher{
		Connection:               connection,
		Channel:                  channel,
		ExchangeName:             rabbitConfig.ExchangeName,
		ThreadCreatedRoutingKey:  rabbitConfig.ThreadCreatedRoutingKey,
		CommentCreatedRoutingKey: rabbitConfig.CommentCreatedRoutingKey,
		ThreadLikedRoutingKey:    rabbitConfig.ThreadLikedRoutingKey,
	}, nil
}
