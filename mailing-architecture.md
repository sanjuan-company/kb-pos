# Mailing Service — Architecture Design

## Overview

A provider-agnostic mailing service built in Go. Designed to support multiple email providers (SendGrid, Mailgun, AWS SES, SMTP, etc.) with zero-code provider swaps via configuration. Follows hexagonal architecture (ports & adapters) with clean separation of concerns.

---

## Core Principles

| Principle | Why |
|-----------|-----|
| **Program to an interface, not an implementation** | Swap providers without touching business logic |
| **Provider isolation** | Each provider is its own adapter; a failure in one never leaks into another |
| **Failover & fallback** | Built-in retry chain: primary → secondary → ... → dead-letter queue |
| **Observability by default** | Structured logging, metrics, tracing on every send attempt |
| **Idempotency** | Message dedup via idempotency key so retries never double-send |

---

## Folder Structure

```
mailing-service/
├── cmd/
│   └── server/
│       └── main.go                  # Entrypoint: config load, DI wiring, server start
│
├── internal/
│   ├── config/
│   │   └── config.go                # Env/file-based configuration struct
│   │
│   ├── domain/                      # Core business logic (no external deps)
│   │   ├── email.go                 # Email, Attachment, Address value objects
│   │   ├── provider.go              # Provider interface (port)
│   │   └── service.go               # Orchestrator: sends, retries, fallback logic
│   │
│   ├── port/                        # Secondary ports (driven side)
│   │   ├── provider.go              # Interface definition (duplicated from domain? no — domain owns it)
│   │   └── repository.go            # Interface for persisting email logs / templates
│   │
│   ├── adapter/                     # Adapter implementations (driven side)
│   │   ├── provider/
│   │   │   ├── sendgrid.go          # SendGrid v3/v4 adapter
│   │   │   ├── mailgun.go           # Mailgun adapter
│   │   │   ├── ses.go               # AWS SES adapter
│   │   │   ├── smtp.go              # Generic SMTP adapter
│   │   │   └── mock.go              # Mock provider for tests
│   │   └── repository/
│   │       ├── postgres.go          # PostgreSQL-backed email log storage
│   │       └── redis.go             # Redis-backed idempotency + queue
│   │
│   ├── handler/                     # HTTP handlers (driving side)
│   │   ├── router.go                # Gin router setup
│   │   └── email_handler.go         # POST /send, GET /status/:id, POST /templates
│   │
│   ├── middleware/
│   │   ├── logging.go
│   │   ├── recovery.go
│   │   └── idempotency.go           # Idempotency-key middleware
│   │
│   └── dto/
│       ├── request.go               # Incoming JSON shapes
│       └── response.go              # Outgoing JSON shapes
│
├── pkg/                             # Reusable shared packages
│   ├── logger/
│   │   └── logger.go                # Structured logger wrapper (slog / zap)
│   ├── retry/
│   │   └── retry.go                 # Configurable exponential backoff
│   └── validator/
│       └── validator.go             # Email format, template validation
│
├── migrations/                      # SQL migrations
│   ├── 000001_create_email_logs.up.sql
│   └── 000001_create_email_logs.down.sql
│
├── templates/                       # HTML email templates (Go html/template or similar)
│   ├── welcome.html
│   ├── reset_password.html
│   └── partials/
│       ├── header.html
│       └── footer.html
│
├── docker-compose.yml               # Local dev: app + postgres + redis
├── Dockerfile
├── Makefile
├── go.mod
└── go.sum
```

---

## Architecture Diagram

```
┌──────────────┐     ┌─────────────────────────────────────────────┐
│  HTTP Client  │     │              Mailing Service                │
│  (App/Admin)  │     │                                             │
└──────┬───────┘     │  ┌──────────┐  ┌─────────────────────────┐  │
       │             │  │ Handler  │  │      Domain (Core)      │  │
       │ POST /send  │  │ (Gin)    │──┤                         │  │
       ├────────────►│  │          │  │  ┌───────────────────┐  │  │
       │             │  │          │  │  │   EmailService    │  │  │
       │             │  │          │  │  │  - Send()         │  │  │
       │             │  │          │  │  │  - SendWithTemplate│  │  │
       │             │  │          │  │  │  - RetryLogic()   │  │  │
       │             │  │          │  │  │  - FallbackChain()│  │  │
       │             │  │          │  │  └────────┬──────────┘  │  │
       │             │  └──────────┘  └───────────┼──────────────┘  │
       │             │                             │                │
       │             │              ┌──────────────┴──────────┐    │
       │             │              │     Provider Interface   │    │
       │             │              │  ┌────────────────────┐  │    │
       │             │              │  │ Send(ctx, Email)   │  │    │
       │             │              │  │ error              │  │    │
       │             │              │  └────────────────────┘  │    │
       │             │              └──────────┬───────────────┘    │
       │             │                         │                    │
       │             │         ┌───────────────┼───────────────┐    │
       │             │         │               │               │    │
       │             │   ┌─────▼─────┐  ┌──────▼──────┐  ┌────▼──┐ │
       │             │   │ SendGrid  │  │  Mailgun    │  │ AWS   │ │
       │             │   │ Adapter   │  │  Adapter    │  │ SES   │ │
       │             │   └───────────┘  └─────────────┘  └───────┘ │
       │             │                                             │
       │             │  ┌──────────────────────────────────────┐   │
       │             │  │         Repository Layer              │   │
       │             │  │  ┌──────────┐  ┌────────────────┐    │   │
       │             │  │  │ Postgres │  │    Redis       │    │   │
       │             │  │  │(log/store)│  │(idempotency/  │    │   │
       │             │  │  │          │  │ rate limit/    │    │   │
       │             │  │  │          │  │ template cache)│    │   │
       │             │  │  └──────────┘  └────────────────┘    │   │
       │             │  └──────────────────────────────────────┘   │
       └─────────────┘                                             │
                                                                   └─────► Provider API
                                                                          (SendGrid /
                                                                           Mailgun /
                                                                           SES / SMTP)
```

---

## Core Domain

### Provider Interface (`internal/domain/provider.go`)

```go
type Provider interface {
    Name() string
    Send(ctx context.Context, email *Email) error
    SendBatch(ctx context.Context, emails []*Email) error
    HealthCheck(ctx context.Context) error
}
```

### Email Service (`internal/domain/service.go`)

```go
type EmailService struct {
    primary   Provider
    fallbacks []Provider
    repo      EmailRepository
    retrier   *retry.Retry
}

func (s *EmailService) Send(ctx context.Context, email *Email) error {
    // 1. Idempotency check — skip if already sent
    // 2. Attempt primary
    // 3. On failure → retry with backoff
    // 4. On exhausted retries → fallback chain
    // 5. All failed → dead-letter (persist as failed)
}
```

### Email Value Object (`internal/domain/email.go`)

```go
type Email struct {
    ID            string            // UUID
    IDempotencyKey string           // Client-provided dedup key
    From          Address
    To            []Address
    Cc            []Address
    Bcc           []Address
    Subject       string
    PlainText     string
    HTML          string
    TemplateID    string
    TemplateData  map[string]any
    Attachments   []Attachment
    Headers       map[string]string
    ProviderHint  string            // Optional: force a specific provider
}

type Address struct {
    Email string
    Name  string
}
```

---

## Provider Discovery & Configuration

Providers are registered at startup via a factory. The order of providers in config determines the fallback chain.

```go
type ProviderFactory struct {
    providers map[string]Provider
}

func (f *ProviderFactory) BuildChain(config []ProviderConfig) (Provider, []Provider, error) {
    // config[0] → primary, config[1:] → fallbacks
}
```

**Config example (env vars or YAML):**

```yaml
mailing:
  providers:
    - name: sendgrid
      api_key: ${SENDGRID_API_KEY}
      from_email: no-reply@example.com
      from_name: Example
    - name: mailgun
      domain: mg.example.com
      api_key: ${MAILGUN_API_KEY}
      from_email: no-reply@example.com
    - name: ses
      region: us-east-1
      from_email: no-reply@example.com
    - name: smtp
      host: smtp.example.com
      port: 587
      username: ${SMTP_USER}
      password: ${SMTP_PASS}
  retry:
    max_attempts: 3
    initial_backoff: 1s
    max_backoff: 30s
```

---

## Database Schema (email_logs)

```sql
CREATE TABLE email_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key VARCHAR(255) UNIQUE,
    provider        VARCHAR(50) NOT NULL,
    from_email      VARCHAR(255) NOT NULL,
    to_emails       TEXT[] NOT NULL,
    subject         TEXT NOT NULL,
    body_plain      TEXT,
    body_html       TEXT,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
                              -- pending | sent | failed | bounced | dropped
    provider_message_id VARCHAR(255),
    error_message   TEXT,
    attempts        INT DEFAULT 0,
    sent_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_email_logs_status ON email_logs(status);
CREATE INDEX idx_email_logs_idempotency ON email_logs(idempotency_key);
CREATE INDEX idx_email_logs_created_at ON email_logs(created_at);
```

---

## Data Flow: Sending an Email

```
Client                     Service                       Provider 1             Provider 2
  │                          │                              │                      │
  │  POST /send              │                              │                      │
  │  {to, subject, html}     │                              │                      │
  │ ───────────────────────► │                              │                      │
  │                          │  Validate + build Email      │                      │
  │                          │  Generate ID + idempotency   │                      │
  │                          │  Insert log (pending)        │                      │
  │                          │                              │                      │
  │                          │  ── Send via Provider 1 ────►│                      │
  │                          │                              │                      │
  │                          │  ◄── 200 / error ────────────│                      │
  │                          │                              │                      │
  │                          │  if error (retryable):       │                      │
  │                          │    ── retry (backoff) ──────►│                      │
  │                          │                              │                      │
  │                          │  if exhausted:               │                      │
  │                          │    ── Send via Provider 2 ───┤─────────────────────►│
  │                          │                              │                      │
  │                          │  Update log (sent / failed)  │                      │
  │                          │                              │                      │
  │  ◄── 202 Accepted ───────│                              │                      │
  │  {id, status}            │                              │                      │
```

---

## Key Design Decisions

### 1. Interface-driven providers
`Provider` interface is defined in the domain layer (not in adapter). The domain owns the contract. Adapters implement it. This keeps the core completely unaware of SendGrid, Mailgun, or any external library.

### 2. Fallback chain (not fan-out)
By default, only one provider sends per email. The chain is sequential: primary → fallback₁ → fallback₂. This avoids double-sending. If fan-out (send via all providers) is needed, it becomes a separate `FanOutProvider` wrapper.

### 3. Idempotency first
Every send request must carry an idempotency key (UUID generated by the client or the service). The service checks Redis/DB before sending. If the key exists and the status is `sent`, return the existing result. This makes retries safe.

### 4. Async by default
`POST /send` returns `202 Accepted` with a status URL (`/status/:id`). The actual send happens asynchronously via a background worker or queue. For simpler sync needs, a `?sync=true` query param can force synchronous send.

### 5. Provider health checks
Each provider exposes a `HealthCheck()` method. A periodic background goroutine pings all providers and updates a provider status map. The `/health` endpoint exposes provider health so monitoring knows when fallback is active.

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/send` | Send an email |
| `POST` | `/api/v1/send-batch` | Send batch emails |
| `GET` | `/api/v1/status/:id` | Get email send status |
| `GET` | `/api/v1/health` | Service + provider health |
| `POST` | `/api/v1/templates` | Create email template |
| `GET` | `/api/v1/templates/:id` | Get rendered template |

---

## Testing Strategy

| Layer | Approach |
|-------|----------|
| **Domain** | Unit tests with `mock.Provider` — no external dependencies |
| **Adapters** | Integration tests using [MailHog](https://github.com/mailhog/MailHog) (SMTP) + provider sandbox APIs |
| **Handlers** | `httptest.NewRecorder` with Gin test mode |
| **E2E** | Docker Compose with MailHog capturing all outbound emails |

```go
// Example test: provider chain fallback
func TestFallback(t *testing.T) {
    primary := mock.NewProvider("primary").FailAlways()
    fallback := mock.NewProvider("fallback").SucceedAlways()
    svc := NewEmailService(primary, []Provider{fallback}, ...)

    err := svc.Send(ctx, testEmail())
    assert.NoError(t, err)
    assert.True(t, primary.WasCalled())
    assert.True(t, fallback.WasCalled())
}
```

---

## Future Considerations

- **Webhook receiver** — handle delivery events (bounce, open, click) from providers via `/api/v1/webhooks/:provider`.
- **Template management** — store Go `html/template` strings in DB, render with dynamic data server-side.
- **Queue-backed sending** — offload sends to RabbitMQ / Redis streams / SQS for durability and rate limiting.
- **Attachment store** — S3/MinIO integration for large attachments; send URL references instead of base64.
- **Rate limiting** — per-provider rate limiter (token bucket) to stay within API quotas.
- **OpenTelemetry** — trace each send hop through the fallback chain.
