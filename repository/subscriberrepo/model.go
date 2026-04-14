package subscriberrepo

import amqp "github.com/rabbitmq/amqp091-go"

type SubscriberRepository struct {
	channel *amqp.Channel
}

func NewSubscriberRepository(channel *amqp.Channel) *SubscriberRepository {
	return &SubscriberRepository{
		channel: channel,
	}
}
