package controller

import (
	controllerhealth "kaktus-consumer/controller/health"
	controllerthread "kaktus-consumer/controller/thread"
)

type HealthHandler = controllerhealth.Handler
type HealthControllerDependency = controllerhealth.Dependency

type ThreadConsumerHandler = controllerthread.Handler
type ThreadConsumerControllerDependency = controllerthread.Dependency

func NewHealthController(dependency HealthControllerDependency) controllerhealth.Controller {
	return controllerhealth.NewController(dependency)
}

func NewThreadConsumerController(
	dependency ThreadConsumerControllerDependency,
) controllerthread.Controller {
	return controllerthread.NewController(dependency)
}
