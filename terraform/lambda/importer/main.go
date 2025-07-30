package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	importerConfig "github.com/m-mizutani/seccamp-2025-b1/terraform/lambda/importer/config"
	"github.com/m-mizutani/seccamp-2025-b1/terraform/lambda/importer/client"
	"github.com/m-mizutani/seccamp-2025-b1/terraform/lambda/importer/transformer"
	"github.com/m-mizutani/seccamp-2025-b1/terraform/lambda/importer/uploader"
)

type ImporterHandler struct {
	config      *importerConfig.Config
	auditClient *client.AuditlogClient
	transformer *transformer.JSONLTransformer
	uploader    *uploader.S3Uploader
}

func NewImporterHandler() (*ImporterHandler, error) {
	// Load configuration
	cfg, err := importerConfig.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Initialize AWS SDK
	awsConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Initialize clients
	auditClient := client.NewAuditlogClient(cfg.AuditlogURL, cfg.Timeout())
	transformer := transformer.NewJSONLTransformer()
	s3Client := s3.NewFromConfig(awsConfig)
	uploader := uploader.NewS3Uploader(s3Client, cfg.S3BucketName, cfg.AWSRegion)

	return &ImporterHandler{
		config:      cfg,
		auditClient: auditClient,
		transformer: transformer,
		uploader:    uploader,
	}, nil
}

func (h *ImporterHandler) Handle(ctx context.Context, event events.EventBridgeEvent) error {
	startTime := time.Now()
	log.Printf("Importer Lambda started at %s", startTime.Format("2006-01-02T15:04:05Z"))

	// Calculate time range for log fetching
	timeRange := CalculateTimeRange(startTime, h.config.BufferMinutes)
	log.Printf("Fetching logs for time range: %s", timeRange.String())

	// Fetch all logs from auditlog API
	logs, err := h.auditClient.FetchAllLogs(ctx, timeRange.StartTime, timeRange.EndTime)
	if err != nil {
		return fmt.Errorf("failed to fetch logs: %w", err)
	}

	log.Printf("Fetched %d log entries", len(logs))

	// Skip processing if no logs
	if len(logs) == 0 {
		log.Println("No logs to process, skipping upload")
		return nil
	}

	// Transform logs to compressed JSONL format
	compressedData, err := h.transformer.Transform(logs)
	if err != nil {
		return fmt.Errorf("failed to transform logs: %w", err)
	}

	log.Printf("Transformed and compressed %d logs to %d bytes", len(logs), len(compressedData))

	// Upload to S3
	key, err := h.uploader.UploadWithTimestamp(ctx, compressedData, startTime)
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	duration := time.Since(startTime)
	log.Printf("Successfully uploaded logs to S3 key: %s", key)
	log.Printf("Importer Lambda completed successfully in %v", duration)

	return nil
}

func main() {
	handler, err := NewImporterHandler()
	if err != nil {
		log.Fatalf("Failed to create handler: %v", err)
	}

	lambda.Start(handler.Handle)
}