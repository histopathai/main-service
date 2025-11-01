package config

import (
	"fmt"
	"os"
	"time"
)

type Environment string

const (
	EnvLocal      Environment = "LOCAL"
	EnvDev        Environment = "DEV"
	EnvProduction Environment = "PROD"
)

type Config struct {
	Env     Environment
	Server  ServerConfig
	GCP     GCPConfig
	PubSub  PubSubConfig
	Logging LoggingConfig
}

type ServerConfig struct {
	Port         string
	GinMode      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type GCPConfig struct {
	ProjectID           string
	CredentialsFile     string
	OriginalBucketName  string
	ProcessedBucketName string
}

type PubSubConfig struct {
	UploadStatusTopicID              string
	ImageProcessingTopicID           string
	ImageProcessResultTopicID        string
	UploadStatusSubscriptionID       string
	ImageProcessResultSubscriptionID string
}

type LoggingConfig struct {
	Level  string
	Format string // "json" or "text"
}

func Load() (*Config, error) {
	env := getEnv("ENV", "LOCAL")

	readTimeout, err := time.ParseDuration(getEnv("READ_TIMEOUT", "15s"))
	if err != nil {
		return nil, fmt.Errorf("invalid READ_TIMEOUT: %w", err)
	}

	writeTimeout, err := time.ParseDuration(getEnv("WRITE_TIMEOUT", "30s"))
	if err != nil {
		return nil, fmt.Errorf("invalid WRITE_TIMEOUT: %w", err)
	}

	idleTimeout, err := time.ParseDuration(getEnv("IDLE_TIMEOUT", "120s"))
	if err != nil {
		return nil, fmt.Errorf("invalid IDLE_TIMEOUT: %w", err)
	}

	cfg := &Config{
		Env: Environment(env),
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			GinMode:      getEnv("GIN_MODE", "debug"),
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
		GCP: GCPConfig{
			ProjectID:           requireEnv("PROJECT_ID"),
			CredentialsFile:     getEnv("GOOGLE_APPLICATION_CREDENTIALS", ""),
			OriginalBucketName:  requireEnv("ORIGINAL_BUCKET_NAME"),
			ProcessedBucketName: getEnv("PROCESSED_BUCKET_NAME", ""),
		},
		PubSub: PubSubConfig{
			UploadStatusTopicID:              getEnv("UPLOAD_STATUS_TOPIC_ID", "upload-status"),
			ImageProcessingTopicID:           getEnv("IMAGE_PROCESSING_TOPIC_ID", "image-processing"),
			ImageProcessResultTopicID:        getEnv("IMAGE_PROCESS_RESULT_TOPIC_ID", "image-process-result"),
			UploadStatusSubscriptionID:       getEnv("UPLOAD_STATUS_SUBSCRIPTION_ID", "upload-status-sub"),
			ImageProcessResultSubscriptionID: getEnv("IMAGE_PROCESS_RESULT_SUBSCRIPTION_ID", "image-process-result-sub"),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.GCP.ProjectID == "" {
		return fmt.Errorf("PROJECT_ID is required")
	}

	if c.GCP.OriginalBucketName == "" {
		return fmt.Errorf("ORIGINAL_BUCKET_NAME is required")
	}

	if c.Server.Port == "" {
		return fmt.Errorf("PORT is required")
	}

	return nil
}

func (c *Config) IsProduction() bool {
	return c.Env == EnvProduction
}

func (c *Config) IsDevelopment() bool {
	return c.Env == EnvDev
}

func (c *Config) IsLocal() bool {
	return c.Env == EnvLocal
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func requireEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return value
}
