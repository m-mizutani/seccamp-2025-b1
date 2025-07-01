package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

	return &Handler{
		s3Client:           s3.NewFromConfig(cfg),
		securityLakeBucket: securityLakeBucket,
		region:             region,
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

	// Parse JSONL file
	var rawLogs []RawLog
	scanner := bufio.NewScanner(resp.Body)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var rawLog RawLog
		if err := json.Unmarshal([]byte(line), &rawLog); err != nil {
			log.Printf("Failed to parse line %d: %s, error: %v", lineNum, line, err)
			continue
		}
		rawLogs = append(rawLogs, rawLog)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	if len(rawLogs) == 0 {
		log.Printf("No valid logs found in file %s", key)
		return nil
	}

	// Convert to Parquet format
	var parquetLogs []ParquetLog
	for _, rawLog := range rawLogs {
		parquetLogs = append(parquetLogs, rawLog.ToParquetLog())
	}

	// Generate Parquet file
	parquetData, err := h.generateParquetFile(parquetLogs)
	if err != nil {
		return fmt.Errorf("failed to generate parquet file: %w", err)
	}

	// Generate Security Lake compliant path
	// Security Lake path format: ext/{source_name}/region={region}/accountId={accountId}/eventDay={YYYYMMDD}/eventHour={HH}/
	now := time.Now().UTC()
	accountId := os.Getenv("AWS_ACCOUNT_ID")
	if accountId == "" {
		return fmt.Errorf("AWS_ACCOUNT_ID environment variable is required")
	}

	securityLakeKey := fmt.Sprintf("ext/service-logs/region=%s/accountId=%s/eventDay=%s/eventHour=%02d/%s.parquet",
		h.region,
		accountId,
		now.Format("20060102"),
		now.Hour(),
		strings.ReplaceAll(key, "/", "_"))

	// Upload to Security Lake S3 bucket
	_, err = h.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(h.securityLakeBucket),
		Key:         aws.String(securityLakeKey),
		Body:        bytes.NewReader(parquetData),
		ContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload parquet file to Security Lake: %w", err)
	}

	log.Printf("Successfully uploaded parquet file to Security Lake: s3://%s/%s", h.securityLakeBucket, securityLakeKey)
	return nil
}

func (h *Handler) generateParquetFile(logs []ParquetLog) ([]byte, error) {
	var buf bytes.Buffer

	writer := parquet.NewGenericWriter[ParquetLog](&buf)
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
