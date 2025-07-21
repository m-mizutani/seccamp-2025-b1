package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/parquet-go/parquet-go"
)

type Handler struct {
	s3Client           S3API
	securityLakeBucket string
	region             string
	customLogSource    string
}

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))
}

func NewHandler() (*Handler, error) {
	slog.Info("Initializing converter handler")
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		slog.Error("Failed to load AWS config", "error", err)
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "ap-northeast-1" // default region
		slog.Warn("AWS_REGION not set, using default", "region", region)
	} else {
		slog.Info("AWS_REGION configured", "region", region)
	}

	securityLakeBucket := os.Getenv("SECURITY_LAKE_BUCKET")
	slog.Info("Security Lake bucket configured", "bucket", securityLakeBucket)
	if securityLakeBucket == "" {
		slog.Error("SECURITY_LAKE_BUCKET environment variable is not set")
		return nil, fmt.Errorf("SECURITY_LAKE_BUCKET environment variable is required")
	}

	customLogSource := os.Getenv("CUSTOM_LOG_SOURCE")
	if customLogSource == "" {
		customLogSource = "google-workspace" // default value
		slog.Warn("CUSTOM_LOG_SOURCE not set, using default", "source", customLogSource)
	} else {
		slog.Info("Custom log source configured", "source", customLogSource)
	}

	handler := &Handler{
		s3Client:           s3.NewFromConfig(cfg),
		securityLakeBucket: securityLakeBucket,
		region:             region,
		customLogSource:    customLogSource,
	}
	slog.Info("Converter handler initialized successfully")
	return handler, nil
}

func (h *Handler) HandleSQSEvent(ctx context.Context, event events.SQSEvent) error {
	slog.Info("Received SQS event", "event", event)
	for i, record := range event.Records {
		slog.Info("Processing SQS record", "index", i+1, "total", len(event.Records), "message_id", record.MessageId)
		if err := h.processRecord(ctx, record); err != nil {
			slog.Error("Error processing record", "message_id", record.MessageId, "error", err)
			return err
		}
		slog.Info("Successfully processed SQS record", "index", i+1, "total", len(event.Records))
	}
	slog.Info("Successfully processed all SQS records", "total", len(event.Records))
	return nil
}

func (h *Handler) processRecord(ctx context.Context, record events.SQSMessage) error {
	// Debug: Log the raw SQS message body
	slog.Info("Raw SQS message received", "body", record.Body)

	// Try to detect the message format first
	// Check if it's a direct SNS message or wrapped
	var messageBody string
	var tempMap map[string]interface{}
	if err := json.Unmarshal([]byte(record.Body), &tempMap); err != nil {
		slog.Error("Failed to parse SQS body as JSON", "error", err, "body", record.Body)
		return fmt.Errorf("failed to parse SQS body as JSON: %w", err)
	}

	// Check if this is a wrapped SNS message (SQS receiving from SNS)
	if message, exists := tempMap["Message"]; exists {
		// This is a wrapped SNS message containing an S3 event
		if msgStr, ok := message.(string); ok {
			slog.Info("Found wrapped SNS message, parsing as S3 event", "message", msgStr)
			// Parse directly as S3 event since SNS Message contains S3 event JSON
			var s3Event events.S3Event
			if err := json.Unmarshal([]byte(msgStr), &s3Event); err != nil {
				slog.Error("Failed to parse S3 event from SNS Message", "error", err, "message", msgStr)
				return fmt.Errorf("failed to parse S3 event from SNS Message: %w", err)
			}

			if len(s3Event.Records) == 0 {
				slog.Warn("No S3 records found in SNS message")
				return nil
			}

			slog.Info("Successfully parsed S3 event from SNS", "s3_records", len(s3Event.Records))
			// Process S3 records directly
			for i, s3Record := range s3Event.Records {
				slog.Info("Processing S3 record", "index", i+1, "total", len(s3Event.Records), "event", s3Record.EventName)
				if err := h.processS3Record(ctx, s3Record); err != nil {
					slog.Error("Failed to process S3 record", "index", i+1, "error", err)
					return err
				}
				slog.Info("Successfully processed S3 record", "index", i+1, "total", len(s3Event.Records))
			}
			return nil
		} else {
			slog.Error("Message field is not a string", "type", fmt.Sprintf("%T", message))
			return fmt.Errorf("message field is not a string")
		}
	}

	// This might be a direct SNS event (fallback)
	messageBody = record.Body
	slog.Info("Using direct SQS body as SNS message")

	// Parse SNS message - try different formats
	slog.Info("Parsing SNS message", "message_body", messageBody)

	// First try to parse as SNSEvent (array of records)
	var snsEvent events.SNSEvent
	if err := json.Unmarshal([]byte(messageBody), &snsEvent); err == nil && len(snsEvent.Records) > 0 {
		slog.Info("Parsed as SNSEvent with records", "count", len(snsEvent.Records))
	} else {
		// Try to parse as a single SNS record
		var snsRecord events.SNSEventRecord
		if err := json.Unmarshal([]byte(messageBody), &snsRecord); err == nil {
			slog.Info("Parsed as single SNS record")
			snsEvent.Records = []events.SNSEventRecord{snsRecord}
		} else {
			// Try to parse as direct S3 event (bypassing SNS structure)
			var s3Event events.S3Event
			if err := json.Unmarshal([]byte(messageBody), &s3Event); err == nil && len(s3Event.Records) > 0 {
				slog.Info("Parsed as direct S3 event, bypassing SNS", "s3_records", len(s3Event.Records))
				// Process S3 records directly
				for i, s3Record := range s3Event.Records {
					slog.Info("Processing S3 record directly", "index", i+1, "total", len(s3Event.Records), "event", s3Record.EventName)
					if err := h.processS3Record(ctx, s3Record); err != nil {
						slog.Error("Failed to process S3 record", "index", i+1, "error", err)
						return err
					}
					slog.Info("Successfully processed S3 record", "index", i+1, "total", len(s3Event.Records))
				}
				return nil
			} else {
				slog.Error("Failed to parse message in any format", "sns_error", err, "message_body", messageBody)
				return fmt.Errorf("failed to parse message in any supported format")
			}
		}
	}
	slog.Info("Found SNS records to process", "count", len(snsEvent.Records))

	// Process each SNS record
	for _, snsRecord := range snsEvent.Records {
		if err := h.processSNSRecord(ctx, snsRecord); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) processSNSRecord(ctx context.Context, record events.SNSEventRecord) error {
	// Parse S3 event from SNS message
	slog.Info("Parsing S3 event from SNS message", "subject", record.SNS.Subject)
	var s3Event events.S3Event
	if err := json.Unmarshal([]byte(record.SNS.Message), &s3Event); err != nil {
		slog.Error("Failed to parse S3 event from SNS message", "error", err, "message", record.SNS.Message)
		return fmt.Errorf("failed to parse S3 event: %w", err)
	}
	slog.Info("Found S3 records in SNS message", "count", len(s3Event.Records))

	// Process each S3 record
	for i, s3Record := range s3Event.Records {
		slog.Info("Processing S3 record", "index", i+1, "total", len(s3Event.Records), "event", s3Record.EventName)
		if err := h.processS3Record(ctx, s3Record); err != nil {
			slog.Error("Failed to process S3 record", "index", i+1, "error", err)
			return err
		}
		slog.Info("Successfully processed S3 record", "index", i+1, "total", len(s3Event.Records))
	}

	return nil
}

func (h *Handler) processS3Record(ctx context.Context, record events.S3EventRecord) error {
	bucket := record.S3.Bucket.Name
	key, err := url.QueryUnescape(record.S3.Object.Key)
	if err != nil {
		return fmt.Errorf("failed to decode S3 object key: %w", err)
	}

	slog.Info("Processing S3 object", "bucket", bucket, "key", key)
	slog.Info("Configuration", "security_lake_bucket", h.securityLakeBucket, "region", h.region, "custom_log_source", h.customLogSource)

	// Download the file from S3
	slog.Info("Downloading file from S3", "bucket", bucket, "key", key)
	resp, err := h.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		slog.Error("Failed to get object from S3", "error", err)
		return fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer resp.Body.Close()
	contentLength := int64(0)
	if resp.ContentLength != nil {
		contentLength = *resp.ContentLength
	}
	slog.Info("Successfully downloaded file", "content_length", contentLength)

	// Parse JSON/JSONL file containing Google Workspace logs
	slog.Info("Starting to parse JSON/JSONL file", "file_key", key)
	
	// Check if file is gzip compressed by file extension
	var reader io.Reader = resp.Body
	if strings.HasSuffix(key, ".gz") {
		slog.Info("File is gzip compressed, decompressing", "file_key", key)
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			slog.Error("Failed to create gzip reader", "error", err, "file_key", key)
			return fmt.Errorf("failed to create gzip reader for %s: %w", key, err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}
	
	var gwLogs []GoogleWorkspaceLog
	decoder := json.NewDecoder(reader)
	lineNum := 0

	errorCount := 0
	maxErrors := 100 // Limit consecutive errors to prevent infinite loops
	
	for {
		lineNum++
		var gwLog GoogleWorkspaceLog
		err := decoder.Decode(&gwLog)
		if err == io.EOF {
			slog.Info("Reached end of file", "total_lines", lineNum-1, "parsed_logs", len(gwLogs))
			break
		}
		if err != nil {
			errorCount++
			slog.Warn("Failed to parse JSON at line", "line", lineNum, "error", err, "error_count", errorCount)
			
			// If we hit too many consecutive errors, it's likely the file format is wrong
			if errorCount > maxErrors {
				slog.Error("Too many consecutive JSON parsing errors, stopping", "max_errors", maxErrors, "file_key", key)
				return fmt.Errorf("too many consecutive JSON parsing errors (%d) in file %s", errorCount, key)
			}
			
			// Try to skip to next valid JSON if this is JSONL with a corrupt line
			continue
		}
		
		// Reset error count on successful parse
		errorCount = 0
		gwLogs = append(gwLogs, gwLog)
		
		// Log progress for very large files
		if lineNum%10000 == 0 {
			slog.Info("Processing progress", "lines_processed", lineNum, "logs_parsed", len(gwLogs))
		}
	}

	slog.Info("Parsed Google Workspace logs", "count", len(gwLogs), "file", key)
	if len(gwLogs) == 0 {
		slog.Warn("No valid logs found in file", "file", key)
		return nil
	}

	// Get AWS account ID
	accountID := os.Getenv("AWS_ACCOUNT_ID")
	slog.Info("AWS Account ID", "account_id", accountID)
	if accountID == "" {
		slog.Error("AWS_ACCOUNT_ID environment variable is not set")
		return fmt.Errorf("AWS_ACCOUNT_ID environment variable is required")
	}

	// Convert to OCSF format
	slog.Info("Converting logs to OCSF format", "count", len(gwLogs))
	var ocsfLogs []OCSFWebResourceActivity
	convertedCount := 0
	for i, gwLog := range gwLogs {
		ocsfLog, err := ConvertToOCSF(&gwLog, h.region, accountID)
		if err != nil {
			slog.Error("Failed to convert log to OCSF format", "index", i+1, "error", err)
			continue
		}
		ocsfLogs = append(ocsfLogs, *ocsfLog)
		convertedCount++
	}
	slog.Info("Successfully converted logs to OCSF format", "converted", convertedCount, "total", len(gwLogs))

	if len(ocsfLogs) == 0 {
		slog.Warn("No OCSF logs to process after conversion, skipping file upload")
		return nil
	}

	// Generate Parquet file
	slog.Info("Generating Parquet file", "ocsf_log_count", len(ocsfLogs))
	parquetData, err := h.generateOCSFParquetFile(ocsfLogs)
	if err != nil {
		return fmt.Errorf("failed to generate parquet file: %w", err)
	}
	slog.Info("Generated Parquet file", "size_bytes", len(parquetData))

	// Extract bucket name from ARN if needed
	securityLakeBucket := h.securityLakeBucket
	slog.Info("Processing Security Lake bucket", "original", securityLakeBucket)
	if after, ok := strings.CutPrefix(securityLakeBucket, "arn:aws:s3:::"); ok {
		securityLakeBucket = after
		slog.Info("Extracted bucket name from ARN", "bucket", securityLakeBucket)
	}

	// Generate Security Lake compliant path for custom log source
	// Custom log source path format: ext/{customSourceName}/{version}/region={region}/accountId={accountId}/eventDay={YYYYMMDD}/
	now := time.Now().UTC()
	securityLakeKey := fmt.Sprintf("ext/%s/1.0/region=%s/accountId=%s/eventDay=%s/%s_%s.gz.parquet",
		h.customLogSource,
		h.region,
		accountID,
		now.Format("20060102"),
		now.Format("20060102150405"),
		strings.ReplaceAll(key, "/", "_"))
	slog.Info("Generated Security Lake key", "key", securityLakeKey)

	// Upload to Security Lake S3 bucket
	slog.Info("Uploading to Security Lake S3 bucket", "bucket", securityLakeBucket, "key", securityLakeKey)
	putResp, err := h.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(securityLakeBucket),
		Key:         aws.String(securityLakeKey),
		Body:        bytes.NewReader(parquetData),
		ContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		slog.Error("Failed to upload to Security Lake S3", "error", err, "bucket", securityLakeBucket, "key", securityLakeKey)
		return fmt.Errorf("failed to upload parquet file to Security Lake: %w", err)
	}

	slog.Info("Successfully uploaded parquet file to Security Lake", "bucket", securityLakeBucket, "key", securityLakeKey, "etag", aws.ToString(putResp.ETag))
	return nil
}

func (h *Handler) generateOCSFParquetFile(logs []OCSFWebResourceActivity) ([]byte, error) {
	var buf bytes.Buffer

	writer := parquet.NewGenericWriter[OCSFWebResourceActivity](&buf)
	defer writer.Close()

	_, err := writer.Write(logs)
	if err != nil {
		return nil, fmt.Errorf("failed to write parquet data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close parquet writer: %w", err)
	}

	return buf.Bytes(), nil
}

func main() {
	// Catch any panics during initialization
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Panic during initialization", "panic", r)
			os.Exit(1)
		}
	}()

	slog.Info("Starting converter main function")
	
	handler, err := NewHandler()
	if err != nil {
		slog.Error("Failed to create handler", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting Lambda function")
	
	// Wrap handler with error catching
	wrappedHandler := func(ctx context.Context, event events.SQSEvent) error {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("Panic during event processing", "panic", r)
			}
		}()
		
		slog.Info("Lambda invocation started")
		return handler.HandleSQSEvent(ctx, event)
	}
	
	lambda.Start(wrappedHandler)
}
