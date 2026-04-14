package module

func NewHealthService() HealthService {
	return &healthService{}
}

func (service *healthService) Status() map[string]string {
	return map[string]string{
		"status": "ok",
	}
}
