package service

type Service struct {
	incidents   IncidentRepositoryInterface
	coordinates CoordinatesRepositoryInterface
	logger      LoggerInterfaces
}

func NewService(
	incidents IncidentRepositoryInterface,
	coordinates CoordinatesRepositoryInterface,
	logger LoggerInterfaces,
) *Service {
	return &Service{
		incidents:   incidents,
		coordinates: coordinates,
		logger:      logger,
	}
}
