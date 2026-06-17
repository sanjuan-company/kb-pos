package provider

import (
	"fmt"

	"mailing-service/internal/domain"
)

type ProviderConfig struct {
	Name       string `yaml:"name" json:"name"`
	APIToken   string `yaml:"api_token" json:"api_token"`
	Domain     string `yaml:"domain" json:"domain"`
	Region     string `yaml:"region" json:"region"`
	Host       string `yaml:"host" json:"host"`
	Port       int    `yaml:"port" json:"port"`
	Username   string `yaml:"username" json:"username"`
	Password   string `yaml:"password" json:"password"`
	FromEmail  string `yaml:"from_email" json:"from_email"`
	FromName   string `yaml:"from_name" json:"from_name"`
}

type ProviderFactory struct {
	constructors map[string]func(cfg ProviderConfig) (domain.Provider, error)
}

func NewProviderFactory() *ProviderFactory {
	f := &ProviderFactory{
		constructors: make(map[string]func(cfg ProviderConfig) (domain.Provider, error)),
	}
	f.Register("smtp", func(cfg ProviderConfig) (domain.Provider, error) {
		return NewSMTP(SMTPConfig{
			Host:       cfg.Host,
			Port:       cfg.Port,
			Username:   cfg.Username,
			Password:   cfg.Password,
			FromEmail:  cfg.FromEmail,
			FromName:   cfg.FromName,
		}), nil
	})
	return f
}

func (f *ProviderFactory) Register(name string, fn func(cfg ProviderConfig) (domain.Provider, error)) {
	f.constructors[name] = fn
}

func (f *ProviderFactory) Build(configs []ProviderConfig) (domain.Provider, []domain.Provider, error) {
	if len(configs) == 0 {
		return nil, nil, fmt.Errorf("at least one provider config is required")
	}

	primary, err := f.build(configs[0])
	if err != nil {
		return nil, nil, fmt.Errorf("build primary: %w", err)
	}

	var fallbacks []domain.Provider
	for i := 1; i < len(configs); i++ {
		p, err := f.build(configs[i])
		if err != nil {
			return nil, nil, fmt.Errorf("build fallback %d: %w", i, err)
		}
		fallbacks = append(fallbacks, p)
	}

	return primary, fallbacks, nil
}

func (f *ProviderFactory) build(cfg ProviderConfig) (domain.Provider, error) {
	ctor, ok := f.constructors[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("unknown provider type: %s", cfg.Name)
	}
	return ctor(cfg)
}
