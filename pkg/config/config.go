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

type LocalTLSConfig struct {
	CertFile string
	KeyFile  string
}

type GCPConfig struct {
	ProjectID           string
	ProjectNumber       string
	Region              string
	CredentialsFile     string
	OriginalBucketName  string
	ProcessedBucketName string
	FirestoreDatabase   string // Firestore database name (default: "(default)")
}

type PubSubConfig struct {
	ImageProcessingRequest TopicSubscriptionConfig
	ImageProcessingResult  TopicSubscriptionConfig
	ImageDeletion          TopicSubscriptionConfig
	UploadStatus           SubscriptionConfig
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
	Type      string // "cloudrun" or "mock"
	JobSmall  string // EKLENDİ
	JobMedium string // EKLENDİ
	JobLarge  string // EKLENDİ
}

// Config is the main configuration struct
type Config struct {
	Env      Environment
	Server   ServerConfig
	GCP      GCPConfig
	PubSub   PubSubConfig
	Worker   WorkerConfig
	Logging  LoggingConfig
	Retry    RetryConfig
	LocalTLS LocalTLSConfig
}

// RetryConfig defines retry configuration per event type
type RetryConfig struct {
	ImageProcessComplete RetryPolicyConfig
	ImageProcess         RetryPolicyConfig
}

// RetryPolicyConfig defines retry behavior for a specific event type
type RetryPolicyConfig struct {
	MaxAttempts       int
	BaseBackoffMs     int // Base backoff in milliseconds
	MaxBackoffMs      int // Max backoff in milliseconds
	BackoffMultiplier float64
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
			FirestoreDatabase:   getEnv("FIRESTORE_DATABASE", "(default)"),
		},
		PubSub: PubSubConfig{
			UploadStatus: SubscriptionConfig{
				Name:    getEnv("UPLOAD_STATUS_SUBSCRIPTION", "upload-status-sub"),
				Topic:   getEnv("UPLOAD_STATUS_TOPIC", "upload-status"),
				DLQName: getEnv("UPLOAD_STATUS_DLQ", "upload-status-dlq"),
			},
			ImageProcessingRequest: TopicSubscriptionConfig{
				Topic: TopicConfig{
					Name:    getEnv("IMAGE_PROCESSING_REQUEST_TOPIC", "image-processing-request"),
					DLQName: getEnv("IMAGE_PROCESSING_REQUEST_DLQ", "image-processing-request-dlq"),
				},
				Subscription: SubscriptionConfig{
					Name:    getEnv("IMAGE_PROCESSING_REQUEST_SUB", "image-processing-request-sub"),
					Topic:   getEnv("IMAGE_PROCESSING_REQUEST_TOPIC", "image-processing-request"),
					DLQName: getEnv("IMAGE_PROCESSING_REQUEST_SUB_DLQ", "image-processing-request-sub-dlq"),
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
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},

		Worker: WorkerConfig{
			Type:      getEnv("WORKER_TYPE", "cloudrun"),
			JobSmall:  getEnv("CLOUD_RUN_JOB_SMALL", ""),  // EKLENDİ
			JobMedium: getEnv("CLOUD_RUN_JOB_MEDIUM", ""), // EKLENDİ
			JobLarge:  getEnv("CLOUD_RUN_JOB_LARGE", ""),  // EKLENDİ
		},
		Retry: RetryConfig{
			ImageProcessComplete: RetryPolicyConfig{
				MaxAttempts:       getEnvInt("RETRY_IMAGE_PROCESS_COMPLETE_MAX_ATTEMPTS", 5),
				BaseBackoffMs:     getEnvInt("RETRY_IMAGE_PROCESS_COMPLETE_BASE_BACKOFF_MS", 1000),
				MaxBackoffMs:      getEnvInt("RETRY_IMAGE_PROCESS_COMPLETE_MAX_BACKOFF_MS", 60000),
				BackoffMultiplier: 2.0,
			},
			ImageProcess: RetryPolicyConfig{
				MaxAttempts:       getEnvInt("RETRY_IMAGE_PROCESS_MAX_ATTEMPTS", 3),
				BaseBackoffMs:     getEnvInt("RETRY_IMAGE_PROCESS_BASE_BACKOFF_MS", 2000),
				MaxBackoffMs:      getEnvInt("RETRY_IMAGE_PROCESS_MAX_BACKOFF_MS", 30000),
				BackoffMultiplier: 2.0,
			},
		},

		LocalTLS: LocalTLSConfig{
			CertFile: getEnv("CERT_FILE", ""),
			KeyFile:  getEnv("KEY_FILE", ""),
		},
	}

	// Apply environment-based prefixes for dev environment
	if cfg.IsDevelopment() || cfg.IsLocal() {
		cfg.applyDevPrefixes()
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	// GCP Configuration
	if c.GCP.ProjectID == "" {
		return fmt.Errorf("PROJECT_ID is required")
	}
	if c.GCP.Region == "" {
		return fmt.Errorf("REGION is required")
	}
	if c.GCP.OriginalBucketName == "" {
		return fmt.Errorf("ORIGINAL_BUCKET_NAME is required")
	}

	// Server Configuration
	if c.Server.Port == "" {
		return fmt.Errorf("PORT is required")
	}

	// PubSub Configuration
	if c.PubSub.UploadStatus.Name == "" {
		return fmt.Errorf("UPLOAD_STATUS_SUBSCRIPTION is required")
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

	// Worker Configuration
	if c.Worker.Type == "" {
		return fmt.Errorf("WORKER_TYPE is required")
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

// applyDevPrefixes adds "dev-" prefix to all pubsub topics/subscriptions and firestore database for dev environment
func (c *Config) applyDevPrefixes() {
	const devPrefix = "dev-"

	// Apply prefix to Firestore database
	if c.GCP.FirestoreDatabase == "(default)" {
		c.GCP.FirestoreDatabase = devPrefix + "db"
	}

	// Apply prefix to Upload Status
	c.PubSub.UploadStatus.Name = devPrefix + c.PubSub.UploadStatus.Name
	c.PubSub.UploadStatus.Topic = devPrefix + c.PubSub.UploadStatus.Topic
	if c.PubSub.UploadStatus.DLQName != "" {
		c.PubSub.UploadStatus.DLQName = devPrefix + c.PubSub.UploadStatus.DLQName
	}

	// Apply prefix to Image Processing Request
	c.PubSub.ImageProcessingRequest.Topic.Name = devPrefix + c.PubSub.ImageProcessingRequest.Topic.Name
	if c.PubSub.ImageProcessingRequest.Topic.DLQName != "" {
		c.PubSub.ImageProcessingRequest.Topic.DLQName = devPrefix + c.PubSub.ImageProcessingRequest.Topic.DLQName
	}
	c.PubSub.ImageProcessingRequest.Subscription.Name = devPrefix + c.PubSub.ImageProcessingRequest.Subscription.Name
	c.PubSub.ImageProcessingRequest.Subscription.Topic = devPrefix + c.PubSub.ImageProcessingRequest.Subscription.Topic
	if c.PubSub.ImageProcessingRequest.Subscription.DLQName != "" {
		c.PubSub.ImageProcessingRequest.Subscription.DLQName = devPrefix + c.PubSub.ImageProcessingRequest.Subscription.DLQName
	}

	// Apply prefix to Image Processing Result
	c.PubSub.ImageProcessingResult.Topic.Name = devPrefix + c.PubSub.ImageProcessingResult.Topic.Name
	if c.PubSub.ImageProcessingResult.Topic.DLQName != "" {
		c.PubSub.ImageProcessingResult.Topic.DLQName = devPrefix + c.PubSub.ImageProcessingResult.Topic.DLQName
	}
	c.PubSub.ImageProcessingResult.Subscription.Name = devPrefix + c.PubSub.ImageProcessingResult.Subscription.Name
	c.PubSub.ImageProcessingResult.Subscription.Topic = devPrefix + c.PubSub.ImageProcessingResult.Subscription.Topic
	if c.PubSub.ImageProcessingResult.Subscription.DLQName != "" {
		c.PubSub.ImageProcessingResult.Subscription.DLQName = devPrefix + c.PubSub.ImageProcessingResult.Subscription.DLQName
	}

	// Apply prefix to Image Deletion
	c.PubSub.ImageDeletion.Topic.Name = devPrefix + c.PubSub.ImageDeletion.Topic.Name
	if c.PubSub.ImageDeletion.Topic.DLQName != "" {
		c.PubSub.ImageDeletion.Topic.DLQName = devPrefix + c.PubSub.ImageDeletion.Topic.DLQName
	}
	c.PubSub.ImageDeletion.Subscription.Name = devPrefix + c.PubSub.ImageDeletion.Subscription.Name
	c.PubSub.ImageDeletion.Subscription.Topic = devPrefix + c.PubSub.ImageDeletion.Subscription.Topic
	if c.PubSub.ImageDeletion.Subscription.DLQName != "" {
		c.PubSub.ImageDeletion.Subscription.DLQName = devPrefix + c.PubSub.ImageDeletion.Subscription.DLQName
	}

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

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := fmt.Sscanf(value, "%d", new(int)); err == nil && intVal == 1 {
			var result int
			fmt.Sscanf(value, "%d", &result)
			return result
		}
	}
	return defaultValue
}
