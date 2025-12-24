package service

type Service struct {
	incedents IncedentRepositoryInterface
	logger    LoggerInterfaces
}

func NewService(
	incedents IncedentRepositoryInterface,
	logger LoggerInterfaces,
) *Service {
	return &Service{
		incedents: incedents,
		logger:    logger,
	}
}
