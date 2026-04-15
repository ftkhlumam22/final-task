package publisher

import (
	"fmt"
	"sync"

	"kaktus-consumer/model"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PublisherRepository struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	mutex      sync.Mutex
}

func NewPublisherRepository(connection *amqp.Connection) (*PublisherRepository, error) {
	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("%w: open publisher channel: %v", model.ErrConsumeEvent, err)
	}

	return &PublisherRepository{
		connection: connection,
		channel:    channel,
	}, nil
}
