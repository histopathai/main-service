package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ProjectID          string
	OriginalBucketName string
	Server             ServerConfig
	MsgTopics          MsgTopicConfig
	Env                string
}
type MsgTopicConfig struct {
	UploadStatusTopicID       string
	ImageProcessingTopicID    string
	ImageProcessresultTopicID string
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	GINMode      string
}

func LoadConfig() (*Config, error) {
	env := getEnvOrDefault("ENV", "LOCAL")

	if env == "LOCAL" {
		err := godotenv.Load()
		if err != nil {
			slog.Warn("Local .env file not found, proceeding with system environment variables")
		} else {
			slog.Info(".env file loaded successfully")
		}

		gacPath := getEnvOrDefault("GOOGLE_APPLICATION_CREDENTIALS", "")
		if gacPath == "" {
			slog.Warn("GOOGLE_APPLICATION_CREDENTIALS not found, proceeding with default variables")
		} else if _, err := os.Stat(gacPath); os.IsNotExist(err) {
			slog.Warn("GOOGLE_APPLICATION_CREDENTIALS file does not exist", "path", gacPath)
		} else {
			slog.Info("GOOGLE_APPLICATION_CREDENTIALS file found", "path", gacPath)
		}
	}

	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		return nil, fmt.Errorf("PROJECT_ID is required")
	}

	originalBucketName := os.Getenv("ORIGINAL_BUCKET_NAME")
	if originalBucketName == "" {
		return nil, fmt.Errorf("ORIGINAL_BUCKET_NAME is required")
	}

	port := getEnvOrDefault("PORT", "8080")
	readTimeoutStr := getEnvOrDefault("READ_TIMEOUT", "15s")
	readTimeout, err := time.ParseDuration(readTimeoutStr)
	if err != nil {
		slog.Warn("Invalid READ_TIMEOUT value, falling back to default value", "value", readTimeoutStr, "error", err)
		readTimeout = 15 * time.Second
	}

	writeTimeoutStr := getEnvOrDefault("WRITE_TIMEOUT", "60s")
	writeTimeout, err := time.ParseDuration(writeTimeoutStr)
	if err != nil {
		slog.Warn("Invalid WRITE_TIMEOUT value, falling back to default value", "value", writeTimeoutStr, "error", err)
		writeTimeout = 60 * time.Second
	}

	idleTimeoutStr := getEnvOrDefault("IDLE_TIMEOUT", "120s")
	idleTimeout, err := time.ParseDuration(idleTimeoutStr)
	if err != nil {
		slog.Warn("Invalid IDLE_TIMEOUT value, falling back to default value", "value", idleTimeoutStr, "error", err)
		idleTimeout = 120 * time.Second
	}
	ginMode := getEnvOrDefault("GIN_MODE", "debug")
	if env != "LOCAL" {
		ginMode = "release"
	}

	uploadStatusTopicID := os.Getenv("UPLOAD_STATUS_TOPIC_ID")
	if uploadStatusTopicID == "" {
		return nil, fmt.Errorf("UPLOAD_STATUS_TOPIC_ID is required")
	}

	imageProcessingTopicID := os.Getenv("IMAGE_PROCESSING_TOPIC_ID")
	if imageProcessingTopicID == "" {
		return nil, fmt.Errorf("IMAGE_PROCESSING_TOPIC_ID is required")
	}

	ImageProcessresultTopicID := os.Getenv("IMAGE_PROCESS_STATUS_TOPIC_ID")
	if ImageProcessresultTopicID == "" {
		return nil, fmt.Errorf("IMAGE_PROCESS_STATUS_TOPIC_ID is required")
	}

	msgTopics := MsgTopicConfig{
		UploadStatusTopicID:       uploadStatusTopicID,
		ImageProcessingTopicID:    imageProcessingTopicID,
		ImageProcessresultTopicID: ImageProcessresultTopicID,
	}

	config := &Config{
		ProjectID:          projectID,
		OriginalBucketName: originalBucketName,
		Env:                env,
		Server: ServerConfig{
			Port:         port,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
			GINMode:      ginMode,
		},
		MsgTopics: msgTopics,
	}

	return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	slog.Debug(fmt.Sprintf("Environment variable %s not set, using default value: %s", key, defaultValue))
	return defaultValue
}
