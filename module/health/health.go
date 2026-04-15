package health

func NewService() Service {
	return &service{}
}

func (service *service) Status() map[string]string {
	return map[string]string{
		"status": "ok",
	}
}
