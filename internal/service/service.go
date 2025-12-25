package service

type Service struct {
	incidents IncedentRepositoryInterface
	logger    LoggerInterfaces
}

func NewService(
	incidents IncedentRepositoryInterface,
	logger LoggerInterfaces,
) *Service {
	return &Service{
		incidents: incidents,
		logger:    logger,
	}
}
