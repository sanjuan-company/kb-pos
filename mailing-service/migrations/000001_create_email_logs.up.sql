CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE email_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key VARCHAR(255) UNIQUE,
    provider VARCHAR(50) NOT NULL,
    from_email VARCHAR(255) NOT NULL,
    from_name VARCHAR(255) DEFAULT '',
    to_emails TEXT[] NOT NULL,
    subject TEXT NOT NULL,
    body_plain TEXT,
    body_html TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    provider_message_id VARCHAR(255),
    error_message TEXT,
    attempts INT DEFAULT 0,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_email_logs_status ON email_logs(status);
CREATE INDEX idx_email_logs_idempotency ON email_logs(idempotency_key);
CREATE INDEX idx_email_logs_created_at ON email_logs(created_at);
