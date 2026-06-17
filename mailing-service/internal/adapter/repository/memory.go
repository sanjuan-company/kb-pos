package repository

import (
	"context"
	"sync"

	"mailing-service/internal/domain"
)

type MemoryStore struct {
	mu    sync.RWMutex
	logs  map[string]*domain.EmailLog
	byKey map[string]*domain.EmailLog
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		logs:  make(map[string]*domain.EmailLog),
		byKey: make(map[string]*domain.EmailLog),
	}
}

func (s *MemoryStore) Save(_ context.Context, log *domain.EmailLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs[log.ID] = log
	if log.IdempotencyKey != "" {
		s.byKey[log.IdempotencyKey] = log
	}
	return nil
}

func (s *MemoryStore) Update(_ context.Context, log *domain.EmailLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs[log.ID] = log
	return nil
}

func (s *MemoryStore) FindByID(_ context.Context, id string) (*domain.EmailLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.logs[id], nil
}

func (s *MemoryStore) FindByIdempotencyKey(_ context.Context, key string) (*domain.EmailLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.byKey[key], nil
}
