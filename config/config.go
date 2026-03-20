package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server         ServerConfig
	Database       DatabaseConfig
	Supabase       SupabaseConfig
	AI             AIConfig
	Upload         UploadConfig
	Log            LogConfig
	Otel           OtelConfig
	AdminBootstrap AdminBootstrapConfig
}

type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	DSN string
}

type SupabaseConfig struct {
	URL               string
	ServiceRoleKey    string
	StorageBucket     string
	InviteRedirectURL string
	AuthUserPath      string
	AdminPath         string
}

type AIConfig struct {
	BaseURL            string
	APIKey             string
	Model              string
	Timeout            time.Duration
	KnowledgeMaxTokens int
	HistoryMaxTokens   int
}

type UploadConfig struct {
	MaxFileBytes   int64
	MaxFilesPerBot int
	AllowedTypes   []string
}

type LogConfig struct {
	Level string
}

type OtelConfig struct {
	ServiceName string
	Enabled     bool
}

type AdminBootstrapConfig struct {
	Email    string
	Password string
}

func Load() (Config, error) {
	cfg := Config{
		Server: ServerConfig{
			Host:            envOrDefault("SERVER_HOST", "0.0.0.0"),
			Port:            intEnvOrDefault("SERVER_PORT", 8080),
			ReadTimeout:     durationEnvOrDefault("SERVER_READ_TIMEOUT", 5*time.Second),
			WriteTimeout:    durationEnvOrDefault("SERVER_WRITE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: durationEnvOrDefault("SERVER_SHUTDOWN_TIMEOUT", 15*time.Second),
		},
		Database: DatabaseConfig{
			DSN: os.Getenv("DATABASE_DSN"),
		},
		Supabase: SupabaseConfig{
			URL:               os.Getenv("SUPABASE_URL"),
			ServiceRoleKey:    os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
			StorageBucket:     envOrDefault("SUPABASE_STORAGE_BUCKET", "chatbot-files"),
			InviteRedirectURL: os.Getenv("SUPABASE_INVITE_REDIRECT_URL"),
			AuthUserPath:      envOrDefault("SUPABASE_AUTH_USER_PATH", "/auth/v1/user"),
			AdminPath:         envOrDefault("SUPABASE_ADMIN_PATH", "/auth/v1/admin"),
		},
		AI: AIConfig{
			BaseURL:            envOrDefault("AI_BASE_URL", "https://api.openai.com/v1"),
			APIKey:             os.Getenv("AI_API_KEY"),
			Model:              envOrDefault("AI_MODEL", "gpt-4.1-mini"),
			Timeout:            durationEnvOrDefault("AI_TIMEOUT", 45*time.Second),
			KnowledgeMaxTokens: intEnvOrDefault("AI_KNOWLEDGE_MAX_TOKENS", 24000),
			HistoryMaxTokens:   intEnvOrDefault("AI_HISTORY_MAX_TOKENS", 12000),
		},
		Upload: UploadConfig{
			MaxFileBytes:   int64EnvOrDefault("UPLOAD_MAX_FILE_BYTES", 10*1024*1024),
			MaxFilesPerBot: intEnvOrDefault("UPLOAD_MAX_FILES_PER_BOT", 5),
			AllowedTypes:   splitCSVOrDefault("UPLOAD_ALLOWED_TYPES", []string{"application/pdf", "text/plain", "text/markdown", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"}),
		},
		Log: LogConfig{
			Level: envOrDefault("LOG_LEVEL", "info"),
		},
		Otel: OtelConfig{
			ServiceName: envOrDefault("OTEL_SERVICE_NAME", "go-chatbot-api"),
			Enabled:     boolEnvOrDefault("OTEL_ENABLED", false),
		},
		AdminBootstrap: AdminBootstrapConfig{
			Email:    os.Getenv("BOOTSTRAP_ADMIN_EMAIL"),
			Password: os.Getenv("BOOTSTRAP_ADMIN_PASSWORD"),
		},
	}

	if cfg.Database.DSN == "" {
		return Config{}, fmt.Errorf("DATABASE_DSN is required")
	}
	if cfg.Supabase.URL == "" {
		return Config{}, fmt.Errorf("SUPABASE_URL is required")
	}
	if cfg.Supabase.ServiceRoleKey == "" {
		return Config{}, fmt.Errorf("SUPABASE_SERVICE_ROLE_KEY is required")
	}
	return cfg, nil
}

func (c Config) ValidateServer() error {
	if c.AI.APIKey == "" {
		return fmt.Errorf("AI_API_KEY is required")
	}
	return nil
}

func (c Config) ValidateBootstrapAdmin() error {
	if c.AdminBootstrap.Email == "" {
		return fmt.Errorf("BOOTSTRAP_ADMIN_EMAIL is required")
	}
	if c.AdminBootstrap.Password == "" {
		return fmt.Errorf("BOOTSTRAP_ADMIN_PASSWORD is required")
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func intEnvOrDefault(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func int64EnvOrDefault(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func durationEnvOrDefault(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func boolEnvOrDefault(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func splitCSVOrDefault(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
