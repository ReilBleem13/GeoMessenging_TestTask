package service

type Service struct {
	incidents   IncidentRepositoryInterface
	coordinates CoordinatesRepositoryInterface
	queue       QueueInterface
	cache       CacheInterface
	logger      LoggerInterfaces
}

func NewService(
	incidents IncidentRepositoryInterface,
	coordinates CoordinatesRepositoryInterface,
	queue QueueInterface,
	cache CacheInterface,
	logger LoggerInterfaces,
) *Service {
	return &Service{
		incidents:   incidents,
		coordinates: coordinates,
		queue:       queue,
		cache:       cache,
		logger:      logger,
	}
}
