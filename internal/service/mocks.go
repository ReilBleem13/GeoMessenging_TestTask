package service

import (
	"context"
	"red_collar/internal/domain"
)

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
type mockCache struct{}

func (m *mockCache) Save(ctx context.Context, data []byte, key string) error {
	return nil
}
func (m *mockCache) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}
func (m *mockCache) Delete(ctx context.Context, key string) (bool, error) {
	return false, nil
}
