package config

import (
	"fmt"
	"os"
)

// Config holds all settings for the service.
// All env-driven settings should end up here.
type Config struct {
	ListenAddr      string // HTTP listen address, e.g. ":8080"
	UpstreamBaseURL string // Upstream LLM base URL, e.g. "https://api.openai.com"
	UpstreamAPIKey  string // API key for the upstream LLM (required)
	ToolDir         string // Directory with tool definition YAML files
	TelemetryPath   string // Path to telemetry JSONL file
}

const (
	envListenAddr      = "LISTEN_ADDR"
	envUpstreamBaseURL = "UPSTREAM_BASE_URL"
	envUpstreamAPIKey  = "UPSTREAM_API_KEY"
	envToolDir         = "TOOL_DIR"
	envTelemetryPath   = "TELEMETRY_PATH"

	defaultListenAddr      = ":8080"
	defaultUpstreamBaseURL = "https://api.openai.com"
	defaultToolDir         = "./tools.d"
	defaultTelemetryPath   = "./runs.jsonl"
)

// Load reads config from env vars, applies defaults, and validates it.
func Load() (Config, error) {
	cfg := Config{
		ListenAddr:      getEnv(envListenAddr, defaultListenAddr),
		UpstreamBaseURL: getEnv(envUpstreamBaseURL, defaultUpstreamBaseURL),
		ToolDir:         getEnv(envToolDir, defaultToolDir),
		TelemetryPath:   getEnv(envTelemetryPath, defaultTelemetryPath),
		UpstreamAPIKey:  os.Getenv(envUpstreamAPIKey),
	}

	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func validate(cfg Config) error {
	if cfg.UpstreamAPIKey == "" {
		return newConfigError(envUpstreamAPIKey, "must be set (no default)")
	}
	if cfg.UpstreamBaseURL == "" {
		return newConfigError(envUpstreamBaseURL, "must not be empty")
	}
	if cfg.ListenAddr == "" {
		return newConfigError(envListenAddr, "must not be empty")
	}
	if cfg.ToolDir == "" {
		return newConfigError(envToolDir, "must not be empty")
	}
	if cfg.TelemetryPath == "" {
		return newConfigError(envTelemetryPath, "must not be empty")
	}
	return nil
}

type ConfigError struct {
	EnvVar string
	Msg    string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("invalid config: %s %s", e.EnvVar, e.Msg)
}

func newConfigError(envVar, msg string) error {
	return &ConfigError{
		EnvVar: envVar,
		Msg:    msg,
	}
}
