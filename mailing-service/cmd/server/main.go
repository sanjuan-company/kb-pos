package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mailing-service/internal/adapter/provider"
	"mailing-service/internal/adapter/repository"
	"mailing-service/internal/config"
	"mailing-service/internal/domain"
	"mailing-service/internal/handler"
	"mailing-service/pkg/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)

	repo := repository.NewMemoryStore()

	smtpProvider := provider.NewSMTP(provider.SMTPConfig{
		Host:       cfg.SMTPHost,
		Port:       cfg.SMTPPort,
		Username:   cfg.SMTPUsername,
		Password:   cfg.SMTPPassword,
		FromEmail:  cfg.FromEmail,
		FromName:   cfg.FromName,
	})

	var fallbackProviders []domain.Provider

	sendgridKey := os.Getenv("SENDGRID_API_KEY")
	if sendgridKey != "" {
		log.Info("sendgrid registered as fallback")
	}

	mailgunKey := os.Getenv("MAILGUN_API_KEY")
	if mailgunKey != "" {
		log.Info("mailgun registered as fallback")
	}

	factory := provider.NewProviderFactory()

	cfgFallbacks := loadFallbackConfigs()
	for _, fbCfg := range cfgFallbacks {
		primary, fbChain, err := factory.Build([]provider.ProviderConfig{fbCfg})
		if err != nil {
			log.Warn("failed to build fallback", "name", fbCfg.Name, "error", err)
			continue
		}
		fallbackProviders = append(fallbackProviders, primary)
		fallbackProviders = append(fallbackProviders, fbChain...)
	}

	svc := domain.NewEmailService(
		smtpProvider,
		fallbackProviders,
		repo,
		cfg.RetryMaxAttempts,
		cfg.RetryInitialBackoff,
		cfg.RetryMaxBackoff,
	)

	router := handler.SetupRouter(svc, repo, log)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("starting server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("forced shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("server stopped")
}

func loadFallbackConfigs() []provider.ProviderConfig {
	var configs []provider.ProviderConfig

	if key := os.Getenv("SENDGRID_API_KEY"); key != "" {
		configs = append(configs, provider.ProviderConfig{
			Name:     "sendgrid",
			APIToken: key,
		})
	}

	if key := os.Getenv("MAILGUN_API_KEY"); key != "" {
		configs = append(configs, provider.ProviderConfig{
			Name:     "mailgun",
			APIToken: key,
			Domain:   os.Getenv("MAILGUN_DOMAIN"),
		})
	}

	return configs
}
