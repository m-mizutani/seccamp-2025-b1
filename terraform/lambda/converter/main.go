package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

func NewHandler() (*Handler, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "ap-northeast-1" // default region
	}

	securityLakeBucket := os.Getenv("SECURITY_LAKE_BUCKET")
	if securityLakeBucket == "" {
		return nil, fmt.Errorf("SECURITY_LAKE_BUCKET environment variable is required")
	}

	customLogSource := os.Getenv("CUSTOM_LOG_SOURCE")
	if customLogSource == "" {
		customLogSource = "google-workspace" // default value
	}

	return &Handler{
		s3Client:           s3.NewFromConfig(cfg),
		securityLakeBucket: securityLakeBucket,
		region:             region,
		customLogSource:    customLogSource,
	}, nil
}

func (h *Handler) HandleSQSEvent(ctx context.Context, event events.SQSEvent) error {
	for _, record := range event.Records {
		if err := h.processRecord(ctx, record); err != nil {
			log.Printf("Error processing record %s: %v", record.MessageId, err)
			return err
		}
	}
	return nil
}

func (h *Handler) processRecord(ctx context.Context, record events.SQSMessage) error {
	// Parse SNS message from SQS
	var snsEvent events.SNSEvent
	if err := json.Unmarshal([]byte(record.Body), &snsEvent); err != nil {
		return fmt.Errorf("failed to parse SNS event: %w", err)
	}

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
	var s3Event events.S3Event
	if err := json.Unmarshal([]byte(record.SNS.Message), &s3Event); err != nil {
		return fmt.Errorf("failed to parse S3 event: %w", err)
	}

	// Process each S3 record
	for _, s3Record := range s3Event.Records {
		if err := h.processS3Record(ctx, s3Record); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) processS3Record(ctx context.Context, record events.S3EventRecord) error {
	bucket := record.S3.Bucket.Name
	key, err := url.QueryUnescape(record.S3.Object.Key)
	if err != nil {
		return fmt.Errorf("failed to decode S3 object key: %w", err)
	}

	log.Printf("Processing S3 object: s3://%s/%s", bucket, key)

	// Download the file from S3
	resp, err := h.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer resp.Body.Close()

	// Parse JSON/JSONL file containing Google Workspace logs
	var gwLogs []GoogleWorkspaceLog
	decoder := json.NewDecoder(resp.Body)
	lineNum := 0

	for {
		lineNum++
		var gwLog GoogleWorkspaceLog
		err := decoder.Decode(&gwLog)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Failed to parse JSON at position %d: %v", lineNum, err)
			// Try to skip to next valid JSON if this is JSONL with a corrupt line
			continue
		}
		gwLogs = append(gwLogs, gwLog)
	}

	if len(gwLogs) == 0 {
		log.Printf("No valid logs found in file %s", key)
		return nil
	}

	// Get AWS account ID
	accountID := os.Getenv("AWS_ACCOUNT_ID")
	if accountID == "" {
		return fmt.Errorf("AWS_ACCOUNT_ID environment variable is required")
	}

	// Convert to OCSF format
	var ocsfLogs []OCSFWebResourceActivity
	for _, gwLog := range gwLogs {
		ocsfLog, err := ConvertToOCSF(&gwLog, h.region, accountID)
		if err != nil {
			log.Printf("Failed to convert log to OCSF format: %v", err)
			continue
		}
		ocsfLogs = append(ocsfLogs, *ocsfLog)
	}

	// Generate Parquet file
	parquetData, err := h.generateOCSFParquetFile(ocsfLogs)
	if err != nil {
		return fmt.Errorf("failed to generate parquet file: %w", err)
	}

	// Extract bucket name from ARN if needed
	securityLakeBucket := h.securityLakeBucket
	if after, ok := strings.CutPrefix(securityLakeBucket, "arn:aws:s3:::"); ok {
		securityLakeBucket = after
	}

	// Generate Security Lake compliant path for custom log source
	// Custom log source path format: ext/{customSourceName}/region={region}/accountId={accountId}/eventDay={YYYYMMDD}/
	now := time.Now().UTC()
	securityLakeKey := fmt.Sprintf("ext/%s/region=%s/accountId=%s/eventDay=%s/%s_%s.gz.parquet",
		h.customLogSource,
		h.region,
		accountID,
		now.Format("20060102"),
		now.Format("20060102150405"),
		strings.ReplaceAll(key, "/", "_"))

	// Upload to Security Lake S3 bucket
	_, err = h.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(securityLakeBucket),
		Key:         aws.String(securityLakeKey),
		Body:        bytes.NewReader(parquetData),
		ContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload parquet file to Security Lake: %w", err)
	}

	log.Printf("Successfully uploaded parquet file to Security Lake: s3://%s/%s", securityLakeBucket, securityLakeKey)
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
	handler, err := NewHandler()
	if err != nil {
		log.Fatal(err)
	}

	lambda.Start(handler.HandleSQSEvent)
}
