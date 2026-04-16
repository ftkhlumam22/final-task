package controller

type Controllers struct {
	Auth   AuthController
	Thread ThreadController
}

func NewControllers(authService AuthService, threadService ThreadService) Controllers {
	return Controllers{
		Auth:   NewAuthController(authService),
		Thread: NewThreadController(threadService),
	}
}
