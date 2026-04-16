package subscriber

import (
	"fmt"

	"kaktus-consumer/model"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (repository SubscriberRepository) Consume(
	queueName string,
	consumerTag string,
) (<-chan amqp.Delivery, error) {
	deliveryChannel, err := repository.channel.Consume(
		queueName,
		consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: start consume queue %s: %v", model.ErrConsumeEvent, queueName, err)
	}

	return deliveryChannel, nil
}

func (repository SubscriberRepository) Ack(deliveryMessage amqp.Delivery) error {
	err := deliveryMessage.Ack(false)
	if err != nil {
		return fmt.Errorf("%w: ack message: %v", model.ErrConsumeEvent, err)
	}

	return nil
}

func (repository SubscriberRepository) Nack(deliveryMessage amqp.Delivery, requeue bool) error {
	err := deliveryMessage.Nack(false, requeue)
	if err != nil {
		return fmt.Errorf("%w: nack message: %v", model.ErrConsumeEvent, err)
	}

	return nil
}
