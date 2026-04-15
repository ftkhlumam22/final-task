package module

import (
	modulehealth "kaktus-consumer/module/health"
	modulethread "kaktus-consumer/module/thread"
)

type ThreadService = modulethread.Service
type ThreadServiceDependency = modulethread.Dependency
type ConsumeDecision = modulethread.ConsumeDecision

type HealthService = modulehealth.Service

func NewThreadService(dependency ThreadServiceDependency) ThreadService {
	return modulethread.NewService(dependency)
}

func NewHealthService() HealthService {
	return modulehealth.NewService()
}
