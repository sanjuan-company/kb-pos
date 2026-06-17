package domain

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type testProvider struct {
	name   string
	failOn int32
	count  atomic.Int32
}

func (p *testProvider) Name() string { return p.name }

func (p *testProvider) Send(_ context.Context, _ *Email) error {
	p.count.Add(1)
	if p.failOn > 0 && p.count.Load() >= p.failOn {
		return nil
	}
	return errors.New("send failed")
}

func (p *testProvider) HealthCheck(_ context.Context) error { return nil }

type testRepo struct {
	logs  map[string]*EmailLog
	byKey map[string]*EmailLog
}

func newTestRepo() *testRepo {
	return &testRepo{
		logs:  make(map[string]*EmailLog),
		byKey: make(map[string]*EmailLog),
	}
}
func (r *testRepo) Save(_ context.Context, l *EmailLog) error {
	r.logs[l.ID] = l
	if l.IdempotencyKey != "" {
		r.byKey[l.IdempotencyKey] = l
	}
	return nil
}
func (r *testRepo) Update(_ context.Context, l *EmailLog) error {
	r.logs[l.ID] = l
	return nil
}
func (r *testRepo) FindByID(_ context.Context, id string) (*EmailLog, error) {
	return r.logs[id], nil
}
func (r *testRepo) FindByIdempotencyKey(_ context.Context, key string) (*EmailLog, error) {
	return r.byKey[key], nil
}

func emailForTest() *Email {
	return &Email{
		ID:      "test-1",
		Subject: "test",
		To:      []Address{{Email: "test@example.com"}},
	}
}

func TestSend_Success_OnPrimary(t *testing.T) {
	primary := &testProvider{name: "primary", failOn: 1}
	svc := NewEmailService(primary, nil, newTestRepo(), 3, time.Millisecond, time.Millisecond)
	_, err := svc.Send(context.Background(), emailForTest())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if primary.count.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", primary.count.Load())
	}
}

func TestSend_Retry_ThenSuccess(t *testing.T) {
	primary := &testProvider{name: "primary", failOn: 3}
	svc := NewEmailService(primary, nil, newTestRepo(), 3, time.Millisecond, time.Millisecond)
	_, err := svc.Send(context.Background(), emailForTest())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if primary.count.Load() != 3 {
		t.Fatalf("expected 3 calls, got %d", primary.count.Load())
	}
}

func TestSend_Fallback(t *testing.T) {
	primary := &testProvider{name: "primary", failOn: -1}
	fallback := &testProvider{name: "fallback", failOn: 1}
	svc := NewEmailService(primary, []Provider{fallback}, newTestRepo(), 2, time.Millisecond, time.Millisecond)
	_, err := svc.Send(context.Background(), emailForTest())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if primary.count.Load() < 2 {
		t.Fatalf("expected primary to retry, got %d calls", primary.count.Load())
	}
	if fallback.count.Load() != 1 {
		t.Fatalf("expected fallback 1 call, got %d", fallback.count.Load())
	}
}

func TestSend_AllFail(t *testing.T) {
	primary := &testProvider{name: "primary", failOn: -1}
	svc := NewEmailService(primary, nil, newTestRepo(), 1, time.Millisecond, time.Millisecond)
	_, err := svc.Send(context.Background(), emailForTest())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSend_Idempotency(t *testing.T) {
	repo := newTestRepo()
	primary := &testProvider{name: "primary", failOn: 1}
	svc := NewEmailService(primary, nil, repo, 3, time.Millisecond, time.Millisecond)

	email := emailForTest()
	email.IdempotencyKey = "dup-key"

	first, err := svc.Send(context.Background(), email)
	if err != nil {
		t.Fatalf("first send failed: %v", err)
	}
	firstCalls := primary.count.Load()

	second, err := svc.Send(context.Background(), email)
	if err != nil {
		t.Fatalf("second send failed: %v", err)
	}
	secondCalls := primary.count.Load()

	if first.ID != second.ID {
		t.Fatal("expected same log on idempotent send")
	}
	if firstCalls != secondCalls {
		t.Fatal("expected no additional provider calls on idempotent send")
	}
}

func TestSend_ContextCancelled(t *testing.T) {
	primary := &testProvider{name: "primary", failOn: -1}
	svc := NewEmailService(primary, nil, newTestRepo(), 3, time.Millisecond, time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.Send(ctx, emailForTest())
	if err == nil {
		t.Fatal("expected context cancelled error")
	}
}
