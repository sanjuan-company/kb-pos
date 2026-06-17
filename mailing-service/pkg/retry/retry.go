package retry

import (
	"context"
	"math"
	"time"
)

type Config struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

type Func func(context.Context) error

func Do(ctx context.Context, cfg Config, fn Func) error {
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := fn(ctx); err != nil {
			lastErr = err
			if attempt < cfg.MaxAttempts-1 {
				backoff := BackoffDuration(cfg.InitialBackoff, cfg.MaxBackoff, attempt)
				timer := time.NewTimer(backoff)
				select {
				case <-timer.C:
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				}
			}
			continue
		}

		return nil
	}

	return lastErr
}

func BackoffDuration(initial, max time.Duration, attempt int) time.Duration {
	wait := float64(initial) * math.Pow(2, float64(attempt))
	if wait > float64(max) {
		wait = float64(max)
	}
	return time.Duration(wait)
}
