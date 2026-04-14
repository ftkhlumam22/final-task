package messaging

import (
	"fmt"

	"kaktus-consumer/model"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	defaultRabbitMQExchange          = "kaktus.events"
	defaultRabbitMQExchangeType      = "topic"
	defaultThreadCreatedRoutingKey   = "thread.created"
	defaultThreadCreatedQueue        = "kaktus.thread.created.queue"
	defaultCommentCreatedRoutingKey  = "comment.created"
	defaultCommentCreatedQueue       = "kaktus.comment.created.queue"
	defaultThreadLikedRoutingKey     = "thread.liked"
	defaultThreadLikedQueue          = "kaktus.thread.liked.queue"
	defaultThreadGetLikedRoutingKey  = "thread.get.liked"
	defaultThreadGetLikedQueue       = "kaktus.thread.get.liked.queue"
)

func LoadRabbitTopologyConfig(environmentFactory model.GetEnvFactory) RabbitTopologyConfig {
	return RabbitTopologyConfig{
		ExchangeName: environmentFactory.GetString("RABBITMQ_EXCHANGE", defaultRabbitMQExchange),
		ExchangeType: environmentFactory.GetString("RABBITMQ_EXCHANGE_TYPE", defaultRabbitMQExchangeType),
		ThreadCreated: QueueBindingConfig{
			RoutingKey: environmentFactory.GetString("RABBITMQ_THREAD_CREATED_KEY", defaultThreadCreatedRoutingKey),
			QueueName:  environmentFactory.GetString("RABBITMQ_THREAD_CREATED_QUEUE", defaultThreadCreatedQueue),
		},
		CommentCreated: QueueBindingConfig{
			RoutingKey: environmentFactory.GetString("RABBITMQ_COMMENT_CREATED_KEY", defaultCommentCreatedRoutingKey),
			QueueName:  environmentFactory.GetString("RABBITMQ_COMMENT_CREATED_QUEUE", defaultCommentCreatedQueue),
		},
		ThreadLiked: QueueBindingConfig{
			RoutingKey: environmentFactory.GetString("RABBITMQ_THREAD_LIKED_KEY", defaultThreadLikedRoutingKey),
			QueueName:  environmentFactory.GetString("RABBITMQ_THREAD_LIKED_QUEUE", defaultThreadLikedQueue),
		},
		ThreadGetLiked: QueueBindingConfig{
			RoutingKey: environmentFactory.GetString("RABBITMQ_THREAD_GET_LIKED_KEY", defaultThreadGetLikedRoutingKey),
			QueueName:  environmentFactory.GetString("RABBITMQ_THREAD_GET_LIKED_QUEUE", defaultThreadGetLikedQueue),
		},
	}
}

func SetupConsumerTopology(channel *amqp.Channel, topologyConfig RabbitTopologyConfig) (ConsumerQueues, error) {
	err := channel.ExchangeDeclare(
		topologyConfig.ExchangeName,
		topologyConfig.ExchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return ConsumerQueues{}, fmt.Errorf("declare exchange: %w", err)
	}

	threadCreatedQueueName, err := declareAndBindQueue(
		channel,
		topologyConfig.ExchangeName,
		topologyConfig.ThreadCreated,
		"thread.created",
	)
	if err != nil {
		return ConsumerQueues{}, err
	}

	commentCreatedQueueName, err := declareAndBindQueue(
		channel,
		topologyConfig.ExchangeName,
		topologyConfig.CommentCreated,
		"comment.created",
	)
	if err != nil {
		return ConsumerQueues{}, err
	}

	threadLikedQueueName, err := declareAndBindQueue(
		channel,
		topologyConfig.ExchangeName,
		topologyConfig.ThreadLiked,
		"thread.liked",
	)
	if err != nil {
		return ConsumerQueues{}, err
	}

	threadGetLikedQueueName, err := declareAndBindQueue(
		channel,
		topologyConfig.ExchangeName,
		topologyConfig.ThreadGetLiked,
		"thread.get.liked",
	)
	if err != nil {
		return ConsumerQueues{}, err
	}

	return ConsumerQueues{
		ThreadCreatedQueueName:  threadCreatedQueueName,
		CommentCreatedQueueName: commentCreatedQueueName,
		ThreadLikedQueueName:    threadLikedQueueName,
		ThreadGetLikedQueueName: threadGetLikedQueueName,
	}, nil
}

func declareAndBindQueue(
	channel *amqp.Channel,
	exchangeName string,
	queueBinding QueueBindingConfig,
	featureName string,
) (string, error) {
	declaredQueue, err := channel.QueueDeclare(
		queueBinding.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("declare queue for %s: %w", featureName, err)
	}

	err = channel.QueueBind(
		declaredQueue.Name,
		queueBinding.RoutingKey,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("bind queue for %s with routing key %s: %w", featureName, queueBinding.RoutingKey, err)
	}

	return declaredQueue.Name, nil
}
