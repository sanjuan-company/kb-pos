package domain

import "context"

type EmailRepository interface {
	Save(ctx context.Context, log *EmailLog) error
	Update(ctx context.Context, log *EmailLog) error
	FindByID(ctx context.Context, id string) (*EmailLog, error)
	FindByIdempotencyKey(ctx context.Context, key string) (*EmailLog, error)
}
