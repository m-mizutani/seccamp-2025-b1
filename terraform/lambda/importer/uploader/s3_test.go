package uploader

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Mock S3 client for testing
type mockS3Client struct {
	putObjectCalls []s3.PutObjectInput
	putObjectError error
}

func (m *mockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	m.putObjectCalls = append(m.putObjectCalls, *params)
	if m.putObjectError != nil {
		return nil, m.putObjectError
	}
	return &s3.PutObjectOutput{}, nil
}

func TestGenerateKey(t *testing.T) {
	uploader := NewS3Uploader(nil, "test-bucket", "ap-northeast-1")
	
	timestamp := time.Date(2024, 8, 12, 10, 5, 30, 0, time.UTC)
	key := uploader.GenerateKey(timestamp)
	
	expected := "2024/08/12/10/import_20240812_100530.jsonl.gz"
	if key != expected {
		t.Errorf("Expected key %s, got %s", expected, key)
	}
}

func TestUpload(t *testing.T) {
	mockClient := &mockS3Client{}
	uploader := NewS3Uploader(mockClient, "test-bucket", "ap-northeast-1")
	
	testData := []byte("test data")
	testKey := "test/key.jsonl.gz"
	
	err := uploader.Upload(context.Background(), testKey, testData)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if len(mockClient.putObjectCalls) != 1 {
		t.Fatalf("Expected 1 PutObject call, got %d", len(mockClient.putObjectCalls))
	}
	
	call := mockClient.putObjectCalls[0]
	if *call.Bucket != "test-bucket" {
		t.Errorf("Expected bucket test-bucket, got %s", *call.Bucket)
	}
	if *call.Key != testKey {
		t.Errorf("Expected key %s, got %s", testKey, *call.Key)
	}
	if *call.ContentType != "application/gzip" {
		t.Errorf("Expected content type application/gzip, got %s", *call.ContentType)
	}
	if *call.ContentLength != int64(len(testData)) {
		t.Errorf("Expected content length %d, got %d", len(testData), *call.ContentLength)
	}
	
	// Read body to verify content
	bodyBytes := make([]byte, len(testData))
	n, err := call.Body.Read(bodyBytes)
	if err != nil && n != len(testData) {
		t.Errorf("Failed to read body: %v", err)
	}
	if !bytes.Equal(bodyBytes, testData) {
		t.Errorf("Body content mismatch")
	}
}

func TestUploadEmptyData(t *testing.T) {
	mockClient := &mockS3Client{}
	uploader := NewS3Uploader(mockClient, "test-bucket", "ap-northeast-1")
	
	err := uploader.Upload(context.Background(), "test/key.jsonl.gz", []byte{})
	if err == nil {
		t.Error("Expected error for empty data")
	}
	
	if len(mockClient.putObjectCalls) != 0 {
		t.Error("Expected no PutObject calls for empty data")
	}
}

func TestUploadError(t *testing.T) {
	mockClient := &mockS3Client{
		putObjectError: fmt.Errorf("S3 error"),
	}
	uploader := NewS3Uploader(mockClient, "test-bucket", "ap-northeast-1")
	
	err := uploader.Upload(context.Background(), "test/key.jsonl.gz", []byte("test"))
	if err == nil {
		t.Error("Expected error from S3 client")
	}
}

func TestUploadWithTimestamp(t *testing.T) {
	mockClient := &mockS3Client{}
	uploader := NewS3Uploader(mockClient, "test-bucket", "ap-northeast-1")
	
	timestamp := time.Date(2024, 8, 12, 10, 5, 30, 0, time.UTC)
	testData := []byte("test data")
	
	key, err := uploader.UploadWithTimestamp(context.Background(), testData, timestamp)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	expectedKey := "2024/08/12/10/import_20240812_100530.jsonl.gz"
	if key != expectedKey {
		t.Errorf("Expected key %s, got %s", expectedKey, key)
	}
	
	if len(mockClient.putObjectCalls) != 1 {
		t.Fatalf("Expected 1 PutObject call, got %d", len(mockClient.putObjectCalls))
	}
	
	call := mockClient.putObjectCalls[0]
	if *call.Key != expectedKey {
		t.Errorf("Expected key %s, got %s", expectedKey, *call.Key)
	}
}