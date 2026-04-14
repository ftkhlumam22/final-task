package publisherrepo

import "final-task/model"

type ThreadEventPublisher struct {
	rabbitPublisher *model.RabbitPublisher
}

type ThreadRPCPublisher struct {
	rabbitRPCClient *model.RabbitRPCClient
}
