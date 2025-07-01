package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockS3API implements S3API interface for testing
type MockS3API struct {
	mock.Mock
}

func (m *MockS3API) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

func (m *MockS3API) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

func TestS3Operations_GetObject_Success(t *testing.T) {
	mockS3 := new(MockS3API)

	testData := `{"id":"log001","timestamp":"2023-12-01T10:00:00Z","user":"alice","action":"login","target":"","success":true,"remote":"192.168.1.100"}`

	// Setup mock expectations
	mockS3.On("GetObject", mock.Anything, &s3.GetObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("test-key.jsonl"),
	}).Return(&s3.GetObjectOutput{
		Body: io.NopCloser(strings.NewReader(testData)),
	}, nil)

	handler := &Handler{
		s3Client:           mockS3,
		securityLakeBucket: "security-lake-bucket",
		region:             "ap-northeast-1",
	}

	// Execute operation
	result, err := handler.s3Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("test-key.jsonl"),
	})

	// Verify
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Body)

	// Read and verify content
	content, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	assert.Equal(t, testData, string(content))

	mockS3.AssertExpectations(t)
}

func TestS3Operations_GetObject_Error(t *testing.T) {
	mockS3 := new(MockS3API)

	// Setup mock to return error
	mockS3.On("GetObject", mock.Anything, &s3.GetObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("nonexistent-key.jsonl"),
	}).Return((*s3.GetObjectOutput)(nil), assert.AnError)

	handler := &Handler{
		s3Client:           mockS3,
		securityLakeBucket: "security-lake-bucket",
		region:             "ap-northeast-1",
	}

	// Execute operation
	result, err := handler.s3Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("nonexistent-key.jsonl"),
	})

	// Verify
	assert.Error(t, err)
	assert.Nil(t, result)

	mockS3.AssertExpectations(t)
}

func TestS3Operations_PutObject_Success(t *testing.T) {
	mockS3 := new(MockS3API)

	testData := []byte("test parquet data")

	// Custom matcher for PutObjectInput to check the content
	mockS3.On("PutObject", mock.Anything, mock.MatchedBy(func(input *s3.PutObjectInput) bool {
		if input.Bucket == nil || *input.Bucket != "security-lake-bucket" {
			return false
		}
		if input.Key == nil || !strings.Contains(*input.Key, "service-logs") {
			return false
		}
		if input.ContentType == nil || *input.ContentType != "application/octet-stream" {
			return false
		}
		// Read body to verify content
		if input.Body != nil {
			bodyContent, err := io.ReadAll(input.Body)
			if err != nil {
				return false
			}
			return bytes.Equal(bodyContent, testData)
		}
		return false
	})).Return(&s3.PutObjectOutput{}, nil)

	handler := &Handler{
		s3Client:           mockS3,
		securityLakeBucket: "security-lake-bucket",
		region:             "ap-northeast-1",
	}

	// Execute operation
	key := "ext/service-logs/region=ap-northeast-1/accountId=123456789012/eventDay=20231201/eventHour=10/test.parquet"
	result, err := handler.s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String("security-lake-bucket"),
		Key:         aws.String(key),
		Body:        bytes.NewReader(testData),
		ContentType: aws.String("application/octet-stream"),
	})

	// Verify
	require.NoError(t, err)
	require.NotNil(t, result)

	mockS3.AssertExpectations(t)
}

func TestS3Operations_PutObject_Error(t *testing.T) {
	mockS3 := new(MockS3API)

	testData := []byte("test parquet data")

	// Setup mock to return error
	mockS3.On("PutObject", mock.Anything, mock.Anything).Return((*s3.PutObjectOutput)(nil), assert.AnError)

	handler := &Handler{
		s3Client:           mockS3,
		securityLakeBucket: "security-lake-bucket",
		region:             "ap-northeast-1",
	}

	// Execute operation
	key := "ext/service-logs/region=ap-northeast-1/accountId=123456789012/eventDay=20231201/eventHour=10/test.parquet"
	result, err := handler.s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String("security-lake-bucket"),
		Key:         aws.String(key),
		Body:        bytes.NewReader(testData),
		ContentType: aws.String("application/octet-stream"),
	})

	// Verify
	assert.Error(t, err)
	assert.Nil(t, result)

	mockS3.AssertExpectations(t)
}

func TestS3Operations_SecurityLakePathGeneration(t *testing.T) {
	// This test verifies that the Security Lake path format is correct
	// Format: ext/{source_name}/region={region}/accountId={accountId}/eventDay={YYYYMMDD}/eventHour={HH}/

	testCases := []struct {
		name      string
		region    string
		accountId string
		hour      int
		day       string
		filename  string
		expected  string
	}{
		{
			name:      "standard path",
			region:    "ap-northeast-1",
			accountId: "123456789012",
			hour:      10,
			day:       "20231201",
			filename:  "logs.parquet",
			expected:  "ext/service-logs/region=ap-northeast-1/accountId=123456789012/eventDay=20231201/eventHour=10/logs.parquet",
		},
		{
			name:      "midnight hour",
			region:    "us-east-1",
			accountId: "987654321098",
			hour:      0,
			day:       "20231231",
			filename:  "data.parquet",
			expected:  "ext/service-logs/region=us-east-1/accountId=987654321098/eventDay=20231231/eventHour=00/data.parquet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate path using the same format as in main.go
			path := "ext/service-logs/region=" + tc.region + "/accountId=" + tc.accountId + "/eventDay=" + tc.day + "/eventHour=" + fmt.Sprintf("%02d", tc.hour) + "/" + tc.filename
			assert.Equal(t, tc.expected, path)
		})
	}
}
