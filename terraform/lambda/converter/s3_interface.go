package main

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3API defines the interface for S3 operations
type S3API interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

// Ensure that s3.Client implements S3API
var _ S3API = (*s3.Client)(nil)

// GetObjectOutput wrapper for testing
type GetObjectResult struct {
	Body io.ReadCloser
}
