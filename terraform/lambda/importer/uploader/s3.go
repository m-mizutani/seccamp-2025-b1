package uploader

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3API interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type S3Uploader struct {
	client     S3API
	bucketName string
	region     string
}

func NewS3Uploader(client S3API, bucketName, region string) *S3Uploader {
	return &S3Uploader{
		client:     client,
		bucketName: bucketName,
		region:     region,
	}
}

// GenerateKey generates S3 key in format: YYYY/MM/DD/HH/import_YYYYMMDD_HHMMSS.jsonl.gz
func (u *S3Uploader) GenerateKey(timestamp time.Time) string {
	return fmt.Sprintf("%04d/%02d/%02d/%02d/import_%04d%02d%02d_%02d%02d%02d.jsonl.gz",
		timestamp.Year(), timestamp.Month(), timestamp.Day(), timestamp.Hour(),
		timestamp.Year(), timestamp.Month(), timestamp.Day(),
		timestamp.Hour(), timestamp.Minute(), timestamp.Second())
}

// Upload uploads compressed data to S3
func (u *S3Uploader) Upload(ctx context.Context, key string, data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("cannot upload empty data")
	}

	input := &s3.PutObjectInput{
		Bucket:        aws.String(u.bucketName),
		Key:           aws.String(key),
		Body:          bytes.NewReader(data),
		ContentType:   aws.String("application/gzip"),
		ContentLength: aws.Int64(int64(len(data))),
	}

	_, err := u.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

// UploadWithTimestamp uploads data with auto-generated timestamp-based key
func (u *S3Uploader) UploadWithTimestamp(ctx context.Context, data []byte, timestamp time.Time) (string, error) {
	key := u.GenerateKey(timestamp)
	
	if err := u.Upload(ctx, key, data); err != nil {
		return "", fmt.Errorf("failed to upload with key %s: %w", key, err)
	}

	return key, nil
}