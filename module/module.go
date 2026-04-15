package module

import (
	"final-task/model"
	authmodule "final-task/module/auth"
	threadmodule "final-task/module/thread"
	"final-task/repository"
)

type Services struct {
	Auth   *authmodule.AuthService
	Thread *threadmodule.ThreadService
}

func NewServices(
	repositories *repository.Repositories,
	tokenProvider authmodule.TokenProvider,
	rabbitPublisher *model.RabbitPublisher,
	rabbitRPCClient *model.RabbitRPCClient,
) *Services {
	return &Services{
		Auth: authmodule.NewAuthService(
			repositories.Auth.AuthSQL,
			tokenProvider,
			repositories.Auth.AuthCache,
		),
		Thread: threadmodule.NewThreadService(
			repositories.Thread.ThreadSQL,
			repositories.Thread.ThreadCache,
			repositories.Thread.ThreadCache,
			rabbitPublisher,
			rabbitRPCClient,
		),
	}
}
