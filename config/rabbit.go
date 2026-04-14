package config

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func NewRabbitConnection(rabbitURL string) (*amqp.Connection, error) {
	connection, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	return connection, nil
}
