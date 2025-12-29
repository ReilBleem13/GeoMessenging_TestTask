package service

import (
	"context"
	"red_collar/internal/domain"
)

// моки репозитория инцедентов
type mockIncidentsRepository struct {
	createFunc     func(ctx context.Context, incedent *domain.Incident) error
	getByIDFunc    func(ctx context.Context, id int) (*domain.Incident, error)
	paginateFunc   func(ctx context.Context, limit, offset int) ([]domain.Incident, int, error)
	deleteFunc     func(ctx context.Context, id int) error
	fullUpdateFunc func(ctx context.Context, incident *domain.Incident) error
}

func (m *mockIncidentsRepository) Create(ctx context.Context, incedent *domain.Incident) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, incedent)
	}
	return nil
}

func (m *mockIncidentsRepository) GetByID(ctx context.Context, id int) (*domain.Incident, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockIncidentsRepository) Paginate(ctx context.Context, limit, offset int) ([]domain.Incident, int, error) {
	if m.paginateFunc != nil {
		return m.paginateFunc(ctx, limit, offset)
	}
	return nil, 0, nil
}

func (m *mockIncidentsRepository) Delete(ctx context.Context, id int) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockIncidentsRepository) FullUpdate(ctx context.Context, incident *domain.Incident) error {
	if m.fullUpdateFunc != nil {
		return m.fullUpdateFunc(ctx, incident)
	}
	return nil
}

// моки репозитория координат
type mockCoordinatesRepository struct {
	checkFunc    func(ctx context.Context, locCheck *domain.LocationCheck) error
	getStatsFunc func(ctx context.Context, timeWindowsMinutes int) ([]domain.ZoneStat, error)
}

func (m *mockCoordinatesRepository) Check(ctx context.Context, locCheck *domain.LocationCheck) error {
	if m.checkFunc != nil {
		return m.checkFunc(ctx, locCheck)
	}
	return nil
}

func (m *mockCoordinatesRepository) GetStats(ctx context.Context, timeWindowsMinutes int) ([]domain.ZoneStat, error) {
	if m.getStatsFunc != nil {
		return m.getStatsFunc(ctx, timeWindowsMinutes)
	}
	return nil, nil
}

// моки репозитория очереди
type mockQueue struct {
	enqueueFunc func(ctx context.Context, check *domain.LocationCheck) error
}

func (m *mockQueue) Enqueue(ctx context.Context, check *domain.LocationCheck) error {
	if m.enqueueFunc != nil {
		return m.enqueueFunc(ctx, check)
	}
	return nil
}

// моки репозитория логгера
type mockLogger struct {
	infoLogs  []logCall
	errorLogs []logCall
	warnLogs  []logCall
	debugLogs []logCall
}

type logCall struct {
	msg    string
	params []any
}

func (m *mockLogger) Debug(msg string, params ...any) {
	if m == nil {
		return
	}
	m.debugLogs = append(m.debugLogs, logCall{msg: msg, params: params})
}

func (m *mockLogger) Info(msg string, params ...any) {
	if m == nil {
		return
	}
	m.infoLogs = append(m.infoLogs, logCall{msg: msg, params: params})
}

func (m *mockLogger) Warn(msg string, params ...any) {
	if m == nil {
		return
	}
	m.warnLogs = append(m.warnLogs, logCall{msg: msg, params: params})
}

func (m *mockLogger) Error(msg string, params ...any) {
	if m == nil {
		return
	}
	m.errorLogs = append(m.errorLogs, logCall{msg: msg, params: params})
}

func (m *mockLogger) GetInfoLogs() []logCall {
	if m == nil {
		return nil
	}
	return m.infoLogs
}

func (m *mockLogger) GetErrorLogs() []logCall {
	if m == nil {
		return nil
	}
	return m.errorLogs
}

func (m *mockLogger) GetWarnLogs() []logCall {
	if m == nil {
		return nil
	}
	return m.warnLogs
}

// моки репозитория кеша
type mockCache struct {
	saveFunc   func(ctx context.Context, data []byte, key string) error
	getFunc    func(ctx context.Context, key string) ([]byte, error)
	deleteFunc func(ctx context.Context, key string) (bool, error)
}

func (m *mockCache) Save(ctx context.Context, data []byte, key string) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, data, key)
	}
	return nil
}
func (m *mockCache) Get(ctx context.Context, key string) ([]byte, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, key)
	}
	return nil, nil
}
func (m *mockCache) Delete(ctx context.Context, key string) (bool, error) {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, key)
	}
	return false, nil
}
