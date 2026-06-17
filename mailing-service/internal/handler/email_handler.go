package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"mailing-service/internal/domain"
	"mailing-service/internal/dto"
)

type EmailHandler struct {
	svc  *domain.EmailService
	repo domain.EmailRepository
}

func NewEmailHandler(svc *domain.EmailService, repo domain.EmailRepository) *EmailHandler {
	return &EmailHandler{svc: svc, repo: repo}
}

func (h *EmailHandler) Send(c *gin.Context) {
	var req dto.SendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if req.PlainText == "" && req.HTML == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "either plain_text or html is required",
		})
		return
	}

	email := domain.Email{
		ID:             uuid.New().String(),
		IdempotencyKey: req.IdempotencyKey,
		Subject:        req.Subject,
		PlainText:      req.PlainText,
		HTML:           req.HTML,
		ProviderHint:   req.ProviderHint,
	}

	for _, a := range req.To {
		email.To = append(email.To, domain.Address{Email: a.Email, Name: a.Name})
	}
	for _, a := range req.Cc {
		email.Cc = append(email.Cc, domain.Address{Email: a.Email, Name: a.Name})
	}
	for _, a := range req.Bcc {
		email.Bcc = append(email.Bcc, domain.Address{Email: a.Email, Name: a.Name})
	}

	if req.From != nil {
		email.From = domain.Address{Email: req.From.Email, Name: req.From.Name}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	log, err := h.svc.Send(ctx, &email)
	if err != nil {
		c.JSON(http.StatusOK, dto.APIResponse{
			Success: false,
			Message: err.Error(),
			Data: dto.SendResponse{
				ID:       log.ID,
				Status:   log.Status,
				Provider: log.Provider,
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "email processed",
		Data: dto.SendResponse{
			ID:       log.ID,
			Status:   log.Status,
			Provider: log.Provider,
		},
	})
}

func (h *EmailHandler) Status(c *gin.Context) {
	id := c.Param("id")

	log, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil || log == nil {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Message: "email log not found",
		})
		return
	}

	resp := dto.StatusResponse{
		ID:                log.ID,
		Status:            log.Status,
		Provider:          log.Provider,
		ProviderMessageID: log.ProviderMessageID,
		ErrorMessage:      log.ErrorMessage,
		Attempts:          log.Attempts,
		CreatedAt:         log.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         log.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    resp,
	})
}

func (h *EmailHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "ok",
	})
}
