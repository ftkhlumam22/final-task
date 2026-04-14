package messaging

import (
	"fmt"

	"final-task/config"
	"final-task/model"

	amqp "github.com/rabbitmq/amqp091-go"
)

type rabbitClientMode string

const (
	rabbitClientPublisher rabbitClientMode = "publisher"
	rabbitClientRPC       rabbitClientMode = "rpc"
)

func NewRabbitPublisher(rabbitConfig model.RabbitConfig) (*model.RabbitPublisher, error) {
	connection, channel, _, err := newRabbitClient(rabbitConfig, rabbitClientPublisher)
	if err != nil {
		return nil, err
	}

	return &model.RabbitPublisher{
		Connection:   connection,
		Channel:      channel,
		ExchangeName: ExchangeName,
	}, nil
}

func NewRabbitRPCClient(rabbitConfig model.RabbitConfig) (*model.RabbitRPCClient, error) {
	connection, channel, queueName, err := newRabbitClient(rabbitConfig, rabbitClientRPC)
	if err != nil {
		return nil, err
	}

	return &model.RabbitRPCClient{
		Connection:   connection,
		Channel:      channel,
		QueueName:    queueName,
		ExchangeName: ExchangeName,
	}, nil
}

func newRabbitClient(
	rabbitConfig model.RabbitConfig,
	clientMode rabbitClientMode,
) (*amqp.Connection, *amqp.Channel, string, error) {
	connection, err := config.NewRabbitConnection(rabbitConfig.URL)
	if err != nil {
		return nil, nil, "", err
	}

	channel, err := connection.Channel()
	if err != nil {
		_ = connection.Close()
		return nil, nil, "", fmt.Errorf("open rabbitmq channel: %w", err)
	}

	err = declareExchange(channel)
	if err != nil {
		_ = channel.Close()
		_ = connection.Close()
		return nil, nil, "", err
	}

	queueName, err := declareQueueByClientMode(channel, clientMode)
	if err != nil {
		_ = channel.Close()
		_ = connection.Close()
		return nil, nil, "", err
	}

	return connection, channel, queueName, nil
}

func declareExchange(channel *amqp.Channel) error {
	err := channel.ExchangeDeclare(
		ExchangeName,
		ExchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	return nil
}

func declareQueueByClientMode(channel *amqp.Channel, clientMode rabbitClientMode) (string, error) {
	switch clientMode {
	case rabbitClientPublisher:
		return "", nil
	case rabbitClientRPC:
		callbackQueue, err := channel.QueueDeclare(
			"",
			false,
			false,
			true,
			false,
			nil,
		)
		if err != nil {
			return "", fmt.Errorf("declare queue: %w", err)
		}
		return callbackQueue.Name, nil
	default:
		return "", fmt.Errorf("unsupported rabbit client mode: %s", clientMode)
	}
}
