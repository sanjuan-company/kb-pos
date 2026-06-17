package dto

import "mailing-service/internal/domain"

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type SendResponse struct {
	ID       string          `json:"id"`
	Status   domain.EmailStatus `json:"status"`
	Provider string          `json:"provider,omitempty"`
}

type StatusResponse struct {
	ID                string             `json:"id"`
	Status            domain.EmailStatus `json:"status"`
	Provider          string             `json:"provider,omitempty"`
	ProviderMessageID string             `json:"provider_message_id,omitempty"`
	ErrorMessage      string             `json:"error_message,omitempty"`
	Attempts          int                `json:"attempts"`
	CreatedAt         string             `json:"created_at"`
	UpdatedAt         string             `json:"updated_at"`
}
