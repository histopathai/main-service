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

type ServerConfig struct {
	Port         string
	GinMode      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type GCPConfig struct {
	ProjectID           string
	ProjectNumber       string
	Region              string
	CredentialsFile     string
	OriginalBucketName  string
	ProcessedBucketName string
}

type PubSubConfig struct {
	ImageProcessingRequest TopicSubscriptionConfig
	ImageProcessingResult  TopicSubscriptionConfig
	ImageDeletion          TopicSubscriptionConfig
	UploadStatus           SubscriptionConfig
	TelemetryTopic         TopicConfig
	TelemetryDLQ           TopicSubscriptionConfig
	TelemetryError         TopicSubscriptionConfig
}

// TopicSubscriptionConfig bundles topic and subscription together
type TopicSubscriptionConfig struct {
	Topic        TopicConfig
	Subscription SubscriptionConfig
}

type TopicConfig struct {
	Name    string
	DLQName string
}

// SubscriptionConfig represents a PubSub subscription with retry configuration
type SubscriptionConfig struct {
	Name                string
	Topic               string
	DLQName             string
	MaxDeliveryAttempts int
	MinBackoff          time.Duration
	MaxBackoff          time.Duration
	AckDeadline         time.Duration
	RetentionDuration   time.Duration
}
type LoggingConfig struct {
	Level  string
	Format string
}

// WorkerConfig contains worker configuration
type WorkerConfig struct {
	Type       string // "cloudrun" or "mock"
	RunJobName string
}

// Config is the main configuration struct
type Config struct {
	Env     Environment
	Server  ServerConfig
	GCP     GCPConfig
	PubSub  PubSubConfig
	Worker  WorkerConfig
	Logging LoggingConfig
	Retry   RetryConfig
}

// RetryConfig defines global retry configuration
type RetryConfig struct {
	MaxImageProcessingRetries int
	MaxDeletionRetries        int
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
			ProjectNumber:       getEnv("PROJECT_NUMBER", ""),
			Region:              getEnv("REGION", ""),
			CredentialsFile:     getEnv("GOOGLE_APPLICATION_CREDENTIALS", ""),
			OriginalBucketName:  requireEnv("ORIGINAL_BUCKET_NAME"),
			ProcessedBucketName: getEnv("PROCESSED_BUCKET_NAME", ""),
		},
		PubSub: PubSubConfig{
			UploadStatus: SubscriptionConfig{
				Name:    getEnv("UPLOAD_STATUS_SUBSCRIPTION", "upload-status-sub"),
				Topic:   getEnv("UPLOAD_STATUS_TOPIC", "upload-status"),
				DLQName: getEnv("UPLOAD_STATUS_DLQ", "upload-status-dlq"),
			},
			ImageProcessingRequest: TopicSubscriptionConfig{
				Topic: TopicConfig{
					Name:    getEnv("IMAGE_PROCESSING_REQUEST_TOPIC", "image-processing-requests"),
					DLQName: getEnv("IMAGE_PROCESSING_REQUEST_DLQ", "image-processing-requests-dlq"),
				},
				Subscription: SubscriptionConfig{
					Name:    getEnv("IMAGE_PROCESSING_REQUEST_SUB", "image-processing-requests-sub"),
					Topic:   getEnv("IMAGE_PROCESSING_REQUEST_TOPIC", "image-processing-requests"),
					DLQName: getEnv("IMAGE_PROCESSING_REQUEST_SUB_DLQ", "image-processing-requests-sub-dlq"),
				},
			},
			ImageProcessingResult: TopicSubscriptionConfig{
				Topic: TopicConfig{
					Name:    getEnv("IMAGE_PROCESSING_RESULT_TOPIC", "image-processing-results"),
					DLQName: getEnv("IMAGE_PROCESSING_RESULT_DLQ", "image-processing-results-dlq"),
				},
				Subscription: SubscriptionConfig{
					Name:    getEnv("IMAGE_PROCESSING_RESULT_SUB", "image-processing-results-sub"),
					Topic:   getEnv("IMAGE_PROCESSING_RESULT_TOPIC", "image-processing-results"),
					DLQName: getEnv("IMAGE_PROCESSING_RESULT_SUB_DLQ", "image-processing-results-sub-dlq"),
				},
			},
			ImageDeletion: TopicSubscriptionConfig{
				Topic: TopicConfig{
					Name:    getEnv("IMAGE_DELETION_TOPIC", "image-deletion-requests"),
					DLQName: getEnv("IMAGE_DELETION_DLQ", "image-deletion-requests-dlq"),
				},
				Subscription: SubscriptionConfig{
					Name:    getEnv("IMAGE_DELETION_SUB", "image-deletion-requests-sub"),
					Topic:   getEnv("IMAGE_DELETION_TOPIC", "image-deletion-requests"),
					DLQName: getEnv("IMAGE_DELETION_SUB_DLQ", "image-deletion-requests-sub-dlq"),
				},
			},
			TelemetryDLQ: TopicSubscriptionConfig{
				Topic: TopicConfig{
					Name:    getEnv("TELEMETRY_DLQ_TOPIC", "telemetry-dlq"),
					DLQName: "", // DLQ doesn't have its own DLQ
				},
				Subscription: SubscriptionConfig{
					Name:    getEnv("TELEMETRY_DLQ_SUB", "telemetry-dlq-sub"),
					Topic:   getEnv("TELEMETRY_DLQ_TOPIC", "telemetry-dlq"),
					DLQName: "",
				},
			},
			TelemetryError: TopicSubscriptionConfig{
				Topic: TopicConfig{
					Name:    getEnv("TELEMETRY_ERROR_TOPIC", "telemetry-errors"),
					DLQName: "", // Error topic doesn't have DLQ
				},
				Subscription: SubscriptionConfig{
					Name:    getEnv("TELEMETRY_ERROR_SUB", "telemetry-errors-sub"),
					Topic:   getEnv("TELEMETRY_ERROR_TOPIC", "telemetry-errors"),
					DLQName: "",
				},
			},
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Worker: WorkerConfig{
			Type:       getEnv("WORKER_TYPE", "cloudrun"),
			RunJobName: getEnv("CLOUD_RUN_JOB_NAME", ""),
		},
		Retry: RetryConfig{
			MaxImageProcessingRetries: 3,
			MaxDeletionRetries:        3,
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
	if c.GCP.Region == "" {
		return fmt.Errorf("REGION is required")
	}
	if c.PubSub.ImageProcessingRequest.Topic.Name == "" {
		return fmt.Errorf("IMAGE_PROCESSING_REQUEST_TOPIC is required")
	}
	if c.PubSub.ImageProcessingRequest.Subscription.Name == "" {
		return fmt.Errorf("IMAGE_PROCESSING_REQUEST_SUB is required")
	}
	if c.PubSub.ImageProcessingResult.Topic.Name == "" {
		return fmt.Errorf("IMAGE_PROCESSING_RESULT_TOPIC is required")
	}
	if c.PubSub.ImageProcessingResult.Subscription.Name == "" {
		return fmt.Errorf("IMAGE_PROCESSING_RESULT_SUB is required")
	}
	if c.PubSub.ImageDeletion.Topic.Name == "" {
		return fmt.Errorf("IMAGE_DELETION_TOPIC is required")
	}
	if c.PubSub.ImageDeletion.Subscription.Name == "" {
		return fmt.Errorf("IMAGE_DELETION_SUB is required")
	}
	if c.PubSub.TelemetryDLQ.Topic.Name == "" {
		return fmt.Errorf("TELEMETRY_DLQ_TOPIC is required")
	}
	if c.PubSub.TelemetryDLQ.Subscription.Name == "" {
		return fmt.Errorf("TELEMETRY_DLQ_SUB is required")
	}
	if c.PubSub.TelemetryError.Topic.Name == "" {
		return fmt.Errorf("TELEMETRY_ERROR_TOPIC is required")
	}
	if c.PubSub.TelemetryError.Subscription.Name == "" {
		return fmt.Errorf("TELEMETRY_ERROR_SUB is required")
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
