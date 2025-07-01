package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandleSQSEvent_ValidEvent(t *testing.T) {
	// Set required environment variables for test
	os.Setenv("SECURITY_LAKE_BUCKET", "test-security-lake-bucket")
	os.Setenv("AWS_REGION", "ap-northeast-1")
	os.Setenv("AWS_ACCOUNT_ID", "123456789012")
	defer func() {
		os.Unsetenv("SECURITY_LAKE_BUCKET")
		os.Unsetenv("AWS_REGION")
		os.Unsetenv("AWS_ACCOUNT_ID")
	}()

	mockS3 := new(MockS3API)

	// Test data - valid JSONL
	testData := `{"id":"log001","timestamp":"2023-12-01T10:00:00Z","user":"alice","action":"login","target":"","success":true,"remote":"192.168.1.100"}
{"id":"log002","timestamp":"2023-12-01T10:05:00Z","user":"bob","action":"read","target":"document1.txt","success":true,"remote":"192.168.1.101"}`

	// Setup S3 GetObject mock
	mockS3.On("GetObject", mock.Anything, &s3.GetObjectInput{
		Bucket: aws.String("test-raw-logs-bucket"),
		Key:    aws.String("logs/test-file.jsonl"),
	}).Return(&s3.GetObjectOutput{
		Body: io.NopCloser(strings.NewReader(testData)),
	}, nil)

	// Setup S3 PutObject mock - verify Security Lake path format
	mockS3.On("PutObject", mock.Anything, mock.MatchedBy(func(input *s3.PutObjectInput) bool {
		if input.Bucket == nil || *input.Bucket != "test-security-lake-bucket" {
			return false
		}
		if input.Key == nil {
			return false
		}
		key := *input.Key
		// Verify Security Lake path format
		return strings.Contains(key, "ext/service-logs/region=ap-northeast-1/accountId=123456789012/eventDay=") &&
			strings.Contains(key, "eventHour=") &&
			strings.HasSuffix(key, ".parquet")
	})).Return(&s3.PutObjectOutput{}, nil)

	handler := &Handler{
		s3Client:           mockS3,
		securityLakeBucket: "test-security-lake-bucket",
		region:             "ap-northeast-1",
	}

	// Create SQS event with SNS notification containing S3 event
	s3Event := events.S3Event{
		Records: []events.S3EventRecord{
			{
				S3: events.S3Entity{
					Bucket: events.S3Bucket{Name: "test-raw-logs-bucket"},
					Object: events.S3Object{Key: "logs/test-file.jsonl"},
				},
			},
		},
	}

	s3EventJSON, _ := json.Marshal(s3Event)

	snsEvent := events.SNSEvent{
		Records: []events.SNSEventRecord{
			{
				SNS: events.SNSEntity{
					Message: string(s3EventJSON),
				},
			},
		},
	}

	snsEventJSON, _ := json.Marshal(snsEvent)

	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{
				MessageId: "test-message-id",
				Body:      string(snsEventJSON),
			},
		},
	}

	// Execute
	err := handler.HandleSQSEvent(context.Background(), sqsEvent)

	// Verify
	require.NoError(t, err)
	mockS3.AssertExpectations(t)
}

func TestHandleSQSEvent_InvalidJSON(t *testing.T) {
	mockS3 := new(MockS3API)

	handler := &Handler{
		s3Client:           mockS3,
		securityLakeBucket: "test-security-lake-bucket",
		region:             "ap-northeast-1",
	}

	// Create SQS event with invalid JSON
	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{
				MessageId: "test-message-id",
				Body:      "invalid json",
			},
		},
	}

	// Execute
	err := handler.HandleSQSEvent(context.Background(), sqsEvent)

	// Should return error for invalid JSON
	assert.Error(t, err)
	mockS3.AssertExpectations(t)
}

func TestHandleSQSEvent_S3GetObjectError(t *testing.T) {
	// Set required environment variables for test
	os.Setenv("SECURITY_LAKE_BUCKET", "test-security-lake-bucket")
	os.Setenv("AWS_REGION", "ap-northeast-1")
	os.Setenv("AWS_ACCOUNT_ID", "123456789012")
	defer func() {
		os.Unsetenv("SECURITY_LAKE_BUCKET")
		os.Unsetenv("AWS_REGION")
		os.Unsetenv("AWS_ACCOUNT_ID")
	}()

	mockS3 := new(MockS3API)

	// Setup S3 GetObject mock to return error
	mockS3.On("GetObject", mock.Anything, &s3.GetObjectInput{
		Bucket: aws.String("test-raw-logs-bucket"),
		Key:    aws.String("logs/nonexistent-file.jsonl"),
	}).Return((*s3.GetObjectOutput)(nil), assert.AnError)

	handler := &Handler{
		s3Client:           mockS3,
		securityLakeBucket: "test-security-lake-bucket",
		region:             "ap-northeast-1",
	}

	// Create SQS event
	s3Event := events.S3Event{
		Records: []events.S3EventRecord{
			{
				S3: events.S3Entity{
					Bucket: events.S3Bucket{Name: "test-raw-logs-bucket"},
					Object: events.S3Object{Key: "logs/nonexistent-file.jsonl"},
				},
			},
		},
	}

	s3EventJSON, _ := json.Marshal(s3Event)

	snsEvent := events.SNSEvent{
		Records: []events.SNSEventRecord{
			{
				SNS: events.SNSEntity{
					Message: string(s3EventJSON),
				},
			},
		},
	}

	snsEventJSON, _ := json.Marshal(snsEvent)

	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{
				MessageId: "test-message-id",
				Body:      string(snsEventJSON),
			},
		},
	}

	// Execute
	err := handler.HandleSQSEvent(context.Background(), sqsEvent)

	// Should return error when S3 GetObject fails
	assert.Error(t, err)
	mockS3.AssertExpectations(t)
}

func TestHandleSQSEvent_EmptyFile(t *testing.T) {
	// Set required environment variables for test
	os.Setenv("SECURITY_LAKE_BUCKET", "test-security-lake-bucket")
	os.Setenv("AWS_REGION", "ap-northeast-1")
	os.Setenv("AWS_ACCOUNT_ID", "123456789012")
	defer func() {
		os.Unsetenv("SECURITY_LAKE_BUCKET")
		os.Unsetenv("AWS_REGION")
		os.Unsetenv("AWS_ACCOUNT_ID")
	}()

	mockS3 := new(MockS3API)

	// Setup S3 GetObject mock to return empty file
	mockS3.On("GetObject", mock.Anything, &s3.GetObjectInput{
		Bucket: aws.String("test-raw-logs-bucket"),
		Key:    aws.String("logs/empty-file.jsonl"),
	}).Return(&s3.GetObjectOutput{
		Body: io.NopCloser(strings.NewReader("")),
	}, nil)

	// No PutObject should be called for empty file

	handler := &Handler{
		s3Client:           mockS3,
		securityLakeBucket: "test-security-lake-bucket",
		region:             "ap-northeast-1",
	}

	// Create SQS event
	s3Event := events.S3Event{
		Records: []events.S3EventRecord{
			{
				S3: events.S3Entity{
					Bucket: events.S3Bucket{Name: "test-raw-logs-bucket"},
					Object: events.S3Object{Key: "logs/empty-file.jsonl"},
				},
			},
		},
	}

	s3EventJSON, _ := json.Marshal(s3Event)

	snsEvent := events.SNSEvent{
		Records: []events.SNSEventRecord{
			{
				SNS: events.SNSEntity{
					Message: string(s3EventJSON),
				},
			},
		},
	}

	snsEventJSON, _ := json.Marshal(snsEvent)

	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{
				MessageId: "test-message-id",
				Body:      string(snsEventJSON),
			},
		},
	}

	// Execute
	err := handler.HandleSQSEvent(context.Background(), sqsEvent)

	// Should not return error, but also shouldn't call PutObject
	require.NoError(t, err)
	mockS3.AssertExpectations(t)
}

func TestNewHandler_MissingEnvironmentVariables(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("SECURITY_LAKE_BUCKET")

	_, err := NewHandler()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SECURITY_LAKE_BUCKET environment variable is required")
}
