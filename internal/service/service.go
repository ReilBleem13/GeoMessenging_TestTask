package service

type Service struct {
	incidents   IncidentRepositoryInterface
	coordinates CoordinatesRepositoryInterface
	queue       QueueInterface
	logger      LoggerInterfaces
}

func NewService(
	incidents IncidentRepositoryInterface,
	coordinates CoordinatesRepositoryInterface,
	queue QueueInterface,
	logger LoggerInterfaces,
) *Service {
	return &Service{
		incidents:   incidents,
		coordinates: coordinates,
		queue:       queue,
		logger:      logger,
	}
}
