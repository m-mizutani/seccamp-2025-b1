package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AuditlogURL      string
	S3BucketName     string
	AWSRegion        string
	TimeoutSeconds   int
	MaxRetries       int
	BufferMinutes    int
}

func Load() (*Config, error) {
	config := &Config{
		// Default values
		TimeoutSeconds: 240, // 4 minutes
		MaxRetries:     3,
		BufferMinutes:  2, // 7 minutes total (5 + 2)
	}

	// Required environment variables
	config.AuditlogURL = os.Getenv("AUDITLOG_URL")
	if config.AuditlogURL == "" {
		return nil, fmt.Errorf("AUDITLOG_URL environment variable is required")
	}

	config.S3BucketName = os.Getenv("S3_BUCKET_NAME")
	if config.S3BucketName == "" {
		return nil, fmt.Errorf("S3_BUCKET_NAME environment variable is required")
	}

	// AWS_REGION is automatically set by Lambda runtime
	config.AWSRegion = os.Getenv("AWS_REGION")
	if config.AWSRegion == "" {
		config.AWSRegion = "ap-northeast-1" // fallback for local testing
	}

	// Optional environment variables
	if timeoutStr := os.Getenv("TIMEOUT_SECONDS"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil && timeout > 0 {
			config.TimeoutSeconds = timeout
		}
	}

	if retriesStr := os.Getenv("MAX_RETRIES"); retriesStr != "" {
		if retries, err := strconv.Atoi(retriesStr); err == nil && retries >= 0 {
			config.MaxRetries = retries
		}
	}

	if bufferStr := os.Getenv("BUFFER_MINUTES"); bufferStr != "" {
		if buffer, err := strconv.Atoi(bufferStr); err == nil && buffer >= 0 {
			config.BufferMinutes = buffer
		}
	}

	return config, nil
}

func (c *Config) Timeout() time.Duration {
	return time.Duration(c.TimeoutSeconds) * time.Second
}

func (c *Config) BufferDuration() time.Duration {
	return time.Duration(c.BufferMinutes) * time.Minute
}

func (c *Config) Validate() error {
	if c.AuditlogURL == "" {
		return fmt.Errorf("auditlog URL cannot be empty")
	}
	if c.S3BucketName == "" {
		return fmt.Errorf("S3 bucket name cannot be empty")
	}
	if c.TimeoutSeconds <= 0 {
		return fmt.Errorf("timeout seconds must be positive")
	}
	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}
	if c.BufferMinutes < 0 {
		return fmt.Errorf("buffer minutes cannot be negative")
	}
	return nil
}