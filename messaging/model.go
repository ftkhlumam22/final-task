package messaging

type QueueBindingConfig struct {
	RoutingKey string
	QueueName  string
}

type RabbitTopologyConfig struct {
	ExchangeName   string
	ExchangeType   string
	ThreadCreated  QueueBindingConfig
	CommentCreated QueueBindingConfig
	ThreadLiked    QueueBindingConfig
	ThreadGetLiked QueueBindingConfig
}

type ConsumerQueues struct {
	ThreadCreatedQueueName  string
	CommentCreatedQueueName string
	ThreadLikedQueueName    string
	ThreadGetLikedQueueName string
}
