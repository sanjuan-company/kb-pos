package domain

import "time"

type Address struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type Attachment struct {
	Filename string `json:"filename"`
	Content  []byte `json:"content"`
	MIMEType string `json:"mime_type"`
}

type Email struct {
	ID             string       `json:"id"`
	IdempotencyKey string       `json:"idempotency_key,omitempty"`
	From           Address      `json:"from"`
	To             []Address    `json:"to"`
	Cc             []Address    `json:"cc,omitempty"`
	Bcc            []Address    `json:"bcc,omitempty"`
	Subject        string       `json:"subject"`
	PlainText      string       `json:"plain_text,omitempty"`
	HTML           string       `json:"html,omitempty"`
	Attachments    []Attachment `json:"attachments,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	ProviderHint   string       `json:"provider_hint,omitempty"`
}

type EmailStatus string

const (
	StatusPending EmailStatus = "pending"
	StatusSent    EmailStatus = "sent"
	StatusFailed  EmailStatus = "failed"
	StatusBounced EmailStatus = "bounced"
	StatusDropped EmailStatus = "dropped"
)

type EmailLog struct {
	ID                string      `json:"id"`
	IdempotencyKey    string      `json:"idempotency_key,omitempty"`
	Provider          string      `json:"provider"`
	From              Address     `json:"from"`
	To                []Address   `json:"to"`
	Subject           string      `json:"subject"`
	Status            EmailStatus `json:"status"`
	ProviderMessageID string      `json:"provider_message_id,omitempty"`
	ErrorMessage      string      `json:"error_message,omitempty"`
	Attempts          int         `json:"attempts"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}
