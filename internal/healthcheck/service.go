package healthcheck

type ServiceInterface interface {
	Check() []Health
}

type Service struct{}

func NewHealthCheckService() *Service {
	return &Service{}
}

func (s *Service) Check() []Health {
	return []Health{
		{
			ServiceName: ServiceName,
			Status:      "healthy",
			Message:     "Service is running",
		},
	}
}
