package domain

import (
	"context"
	"fmt"
	"math"
	"time"
)

type EmailService struct {
	primary       Provider
	fallbacks     []Provider
	repo          EmailRepository
	maxAttempts   int
	initialBackoff time.Duration
	maxBackoff    time.Duration
}

func NewEmailService(
	primary Provider,
	fallbacks []Provider,
	repo EmailRepository,
	maxAttempts int,
	initialBackoff time.Duration,
	maxBackoff time.Duration,
) *EmailService {
	return &EmailService{
		primary:        primary,
		fallbacks:      fallbacks,
		repo:           repo,
		maxAttempts:    maxAttempts,
		initialBackoff: initialBackoff,
		maxBackoff:     maxBackoff,
	}
}

func (s *EmailService) Send(ctx context.Context, email *Email) (*EmailLog, error) {
	if email.IdempotencyKey != "" {
		existing, err := s.repo.FindByIdempotencyKey(ctx, email.IdempotencyKey)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	log := &EmailLog{
		ID:             email.ID,
		IdempotencyKey: email.IdempotencyKey,
		From:           email.From,
		To:             email.To,
		Subject:        email.Subject,
		Status:         StatusPending,
		Attempts:       0,
		CreatedAt:      time.Now(),
	}

	if err := s.repo.Save(ctx, log); err != nil {
		return nil, fmt.Errorf("save email log: %w", err)
	}

	providers := []Provider{s.primary}
	providers = append(providers, s.fallbacks...)

	providerChain := s.buildProviderChain(providers, email.ProviderHint)

	for _, p := range providerChain {
		for attempt := 0; attempt < s.maxAttempts; attempt++ {
			select {
			case <-ctx.Done():
				log.Status = StatusFailed
				log.ErrorMessage = ctx.Err().Error()
				log.UpdatedAt = time.Now()
				_ = s.repo.Update(ctx, log)
				return log, ctx.Err()
			default:
			}

			err := p.Send(ctx, email)
			log.Attempts++
			log.Provider = p.Name()

			if err == nil {
				log.Status = StatusSent
				log.UpdatedAt = time.Now()
				_ = s.repo.Update(ctx, log)
				return log, nil
			}

			log.ErrorMessage = err.Error()
			_ = s.repo.Update(ctx, log)

			if attempt < s.maxAttempts-1 {
				s.sleepWithContext(ctx, s.backoffDuration(attempt))
			}
		}
	}

	log.Status = StatusFailed
	log.UpdatedAt = time.Now()
	_ = s.repo.Update(ctx, log)

	return log, fmt.Errorf("all providers exhausted, last error: %s", log.ErrorMessage)
}

func (s *EmailService) buildProviderChain(providers []Provider, hint string) []Provider {
	if hint == "" {
		return providers
	}
	for _, p := range providers {
		if p.Name() == hint {
			return []Provider{p}
		}
	}
	return providers
}

func (s *EmailService) backoffDuration(attempt int) time.Duration {
	wait := float64(s.initialBackoff) * math.Pow(2, float64(attempt))
	if wait > float64(s.maxBackoff) {
		wait = float64(s.maxBackoff)
	}
	return time.Duration(wait)
}

func (s *EmailService) sleepWithContext(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ctx.Done():
	}
}
