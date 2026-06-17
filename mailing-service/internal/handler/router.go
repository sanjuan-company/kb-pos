package handler

import (
	"github.com/gin-gonic/gin"

	"mailing-service/internal/domain"
	"mailing-service/internal/middleware"
	"mailing-service/pkg/logger"
)

func SetupRouter(svc *domain.EmailService, repo domain.EmailRepository, log *logger.Logger) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logging(log))

	h := NewEmailHandler(svc, repo)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/send", h.Send)
		v1.GET("/status/:id", h.Status)
		v1.GET("/health", h.Health)
	}

	return r
}
