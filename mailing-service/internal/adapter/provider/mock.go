package provider

import (
	"context"
	"fmt"
	"sync/atomic"

	"mailing-service/internal/domain"
)

var _ domain.Provider = (*MockProvider)(nil)

type MockProvider struct {
	name    string
	fail    atomic.Bool
	calls   atomic.Int64
}

func NewMockProvider(name string) *MockProvider {
	return &MockProvider{name: name}
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Send(_ context.Context, _ *domain.Email) error {
	m.calls.Add(1)
	if m.fail.Load() {
		return fmt.Errorf("%s: simulated failure", m.name)
	}
	return nil
}

func (m *MockProvider) HealthCheck(_ context.Context) error {
	if m.fail.Load() {
		return fmt.Errorf("%s: unhealthy", m.name)
	}
	return nil
}

func (m *MockProvider) Fail()   { m.fail.Store(true) }
func (m *MockProvider) Pass()   { m.fail.Store(false) }
func (m *MockProvider) Calls() int64 { return m.calls.Load() }
