package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// AthenaAPI defines the interface for Athena operations
type AthenaAPI interface {
	StartQueryExecution(ctx context.Context, params *athena.StartQueryExecutionInput, optFns ...func(*athena.Options)) (*athena.StartQueryExecutionOutput, error)
	GetQueryExecution(ctx context.Context, params *athena.GetQueryExecutionInput, optFns ...func(*athena.Options)) (*athena.GetQueryExecutionOutput, error)
	GetQueryResults(ctx context.Context, params *athena.GetQueryResultsInput, optFns ...func(*athena.Options)) (*athena.GetQueryResultsOutput, error)
}

// SNSAPI defines the interface for SNS operations
type SNSAPI interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

// Ensure that the actual clients implement the interfaces
var _ AthenaAPI = (*athena.Client)(nil)
var _ SNSAPI = (*sns.Client)(nil)
