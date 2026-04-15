package health

type Service interface {
	Status() map[string]string
}

type service struct{}
