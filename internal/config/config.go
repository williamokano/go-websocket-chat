package config

import (
	"fmt"
	"net/url"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port           int      `env:"PORT" envDefault:"8080"`
	DatabaseURL    string   `env:"DATABASE_URL" envDefault:"postgres://chat:chat@localhost:5432/chat?sslmode=disable"`
	RedisURL       string   `env:"REDIS_URL" envDefault:"redis://localhost:6379"`
	JWTSecret      string   `env:"JWT_SECRET" envDefault:"dev-secret-change-in-production"`
	RunMigrate     bool     `env:"RUN_MIGRATE" envDefault:"true"`
	AllowedOrigins []string `env:"ALLOWED_ORIGINS" envDefault:"http://localhost:5173,http://localhost:3000" envSeparator:","`

	WebAuthnRPDisplayName string   `env:"WEBAUTHN_RP_DISPLAY_NAME" envDefault:"Chat App"`
	WebAuthnRPID          string   `env:"WEBAUTHN_RP_ID" envDefault:"localhost"`
	WebAuthnRPOrigins     []string `env:"WEBAUTHN_RP_ORIGINS" envDefault:"http://localhost:5173,http://localhost:3000" envSeparator:","`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}

// WebSocketOriginPatterns returns origin patterns suitable for the WebSocket
// library's AcceptOptions. It strips the scheme from each allowed origin,
// returning just "host:port" patterns.
func (c *Config) WebSocketOriginPatterns() []string {
	patterns := make([]string, 0, len(c.AllowedOrigins))
	for _, origin := range c.AllowedOrigins {
		u, err := url.Parse(origin)
		if err != nil {
			patterns = append(patterns, origin)
			continue
		}
		if u.Host != "" {
			patterns = append(patterns, u.Host)
		} else {
			patterns = append(patterns, origin)
		}
	}
	return patterns
}
