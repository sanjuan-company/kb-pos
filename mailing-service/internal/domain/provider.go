package domain

import "context"

type Provider interface {
	Name() string
	Send(ctx context.Context, email *Email) error
	HealthCheck(ctx context.Context) error
}
