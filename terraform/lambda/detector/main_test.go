package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandleEvent_SuccessfulDetection(t *testing.T) {
	// Set required environment variables for test
	os.Setenv("ATHENA_DATABASE", "test-database")
	os.Setenv("ATHENA_RESULTS_BUCKET", "test-results-bucket")
	os.Setenv("ALERTS_SNS_TOPIC_ARN", "arn:aws:sns:ap-northeast-1:123456789012:test-alerts")
	defer func() {
		os.Unsetenv("ATHENA_DATABASE")
		os.Unsetenv("ATHENA_RESULTS_BUCKET")
		os.Unsetenv("ALERTS_SNS_TOPIC_ARN")
	}()

	mockAthena := new(MockAthenaAPI)
	mockSNS := new(MockSNSAPI)

	queryExecutionId := "test-execution-id"

	// Mock StartQueryExecution for all queries
	mockAthena.On("StartQueryExecution", mock.Anything, mock.MatchedBy(func(input *athena.StartQueryExecutionInput) bool {
		return input.QueryString != nil && *input.QueryString != ""
	})).Return(&athena.StartQueryExecutionOutput{
		QueryExecutionId: aws.String(queryExecutionId),
	}, nil).Times(3) // 3 queries will be executed

	// Mock GetQueryExecution - simulate successful completion
	mockAthena.On("GetQueryExecution", mock.Anything, &athena.GetQueryExecutionInput{
		QueryExecutionId: aws.String(queryExecutionId),
	}).Return(&athena.GetQueryExecutionOutput{
		QueryExecution: &types.QueryExecution{
			QueryExecutionId: aws.String(queryExecutionId),
			Status: &types.QueryExecutionStatus{
				State: types.QueryExecutionStateSucceeded,
			},
		},
	}, nil).Times(3)

	// Mock GetQueryResults - return suspicious login detection
	mockAthena.On("GetQueryResults", mock.Anything, &athena.GetQueryResultsInput{
		QueryExecutionId: aws.String(queryExecutionId),
		MaxResults:       aws.Int32(100),
	}).Return(&athena.GetQueryResultsOutput{
		ResultSet: &types.ResultSet{
			Rows: []types.Row{
				// Header row
				{
					Data: []types.Datum{
						{VarCharValue: aws.String("remote")},
						{VarCharValue: aws.String("failed_count")},
						{VarCharValue: aws.String("users")},
					},
				},
				// Data row - suspicious activity detected
				{
					Data: []types.Datum{
						{VarCharValue: aws.String("192.168.1.100")},
						{VarCharValue: aws.String("5")},
						{VarCharValue: aws.String("alice,bob")},
					},
				},
			},
		},
	}, nil).Once().On("GetQueryResults", mock.Anything, &athena.GetQueryResultsInput{
		QueryExecutionId: aws.String(queryExecutionId),
		MaxResults:       aws.Int32(100),
	}).Return(&athena.GetQueryResultsOutput{
		ResultSet: &types.ResultSet{
			Rows: []types.Row{
				// Header row only for other queries (no detection)
				{
					Data: []types.Datum{
						{VarCharValue: aws.String("user")},
						{VarCharValue: aws.String("access_count")},
					},
				},
			},
		},
	}, nil).Times(2)

	// Mock SNS publish - should be called once for the detection
	mockSNS.On("Publish", mock.Anything, mock.MatchedBy(func(input *sns.PublishInput) bool {
		if input.TopicArn == nil || *input.TopicArn != "arn:aws:sns:ap-northeast-1:123456789012:test-alerts" {
			return false
		}

		// Verify alert content
		var alert Alert
		err := json.Unmarshal([]byte(*input.Message), &alert)
		return err == nil && alert.Title == "不審なログイン試行を検知"
	})).Return(&sns.PublishOutput{
		MessageId: aws.String("test-message-id"),
	}, nil).Once()

	handler := &Handler{
		athenaClient:   mockAthena,
		snsClient:      mockSNS,
		database:       "test-database",
		resultsBucket:  "test-results-bucket",
		alertsTopicArn: "arn:aws:sns:ap-northeast-1:123456789012:test-alerts",
		queries: []QueryDefinition{
			{
				Name:        "suspicious_login",
				Description: "同一IPアドレスからの複数回ログイン失敗を検知",
				SQL:         suspiciousLoginQuery,
			},
			{
				Name:        "mass_data_access",
				Description: "短時間での大量データアクセスを検知",
				SQL:         massDataAccessQuery,
			},
			{
				Name:        "failed_auth",
				Description: "複数IPからの継続的な認証失敗を検知",
				SQL:         failedAuthQuery,
			},
		},
	}

	// Create CloudWatch event
	cloudWatchEvent := events.CloudWatchEvent{
		Source: "aws.events",
		Detail: json.RawMessage(`{}`),
	}

	// Execute
	err := handler.HandleEvent(context.Background(), cloudWatchEvent)

	// Verify
	require.NoError(t, err)
	mockAthena.AssertExpectations(t)
	mockSNS.AssertExpectations(t)
}

func TestHandleEvent_NoDetections(t *testing.T) {
	// Set required environment variables for test
	os.Setenv("ATHENA_DATABASE", "test-database")
	os.Setenv("ATHENA_RESULTS_BUCKET", "test-results-bucket")
	os.Setenv("ALERTS_SNS_TOPIC_ARN", "arn:aws:sns:ap-northeast-1:123456789012:test-alerts")
	defer func() {
		os.Unsetenv("ATHENA_DATABASE")
		os.Unsetenv("ATHENA_RESULTS_BUCKET")
		os.Unsetenv("ALERTS_SNS_TOPIC_ARN")
	}()

	mockAthena := new(MockAthenaAPI)
	mockSNS := new(MockSNSAPI)

	queryExecutionId := "test-execution-id"

	// Mock StartQueryExecution
	mockAthena.On("StartQueryExecution", mock.Anything, mock.Anything).Return(&athena.StartQueryExecutionOutput{
		QueryExecutionId: aws.String(queryExecutionId),
	}, nil).Once()

	// Mock GetQueryExecution
	mockAthena.On("GetQueryExecution", mock.Anything, mock.Anything).Return(&athena.GetQueryExecutionOutput{
		QueryExecution: &types.QueryExecution{
			QueryExecutionId: aws.String(queryExecutionId),
			Status: &types.QueryExecutionStatus{
				State: types.QueryExecutionStateSucceeded,
			},
		},
	}, nil).Once()

	// Mock GetQueryResults - return empty results (no detections)
	mockAthena.On("GetQueryResults", mock.Anything, mock.Anything).Return(&athena.GetQueryResultsOutput{
		ResultSet: &types.ResultSet{
			Rows: []types.Row{
				// Header row only
				{
					Data: []types.Datum{
						{VarCharValue: aws.String("column1")},
						{VarCharValue: aws.String("column2")},
					},
				},
			},
		},
	}, nil).Once()

	// SNS should not be called since no detections
	// No SNS mock expectations set

	handler := &Handler{
		athenaClient:   mockAthena,
		snsClient:      mockSNS,
		database:       "test-database",
		resultsBucket:  "test-results-bucket",
		alertsTopicArn: "arn:aws:sns:ap-northeast-1:123456789012:test-alerts",
		queries: []QueryDefinition{
			{Name: "test_query", Description: "Test", SQL: "SELECT 1"},
		},
	}

	cloudWatchEvent := events.CloudWatchEvent{
		Source: "aws.events",
		Detail: json.RawMessage(`{}`),
	}

	err := handler.HandleEvent(context.Background(), cloudWatchEvent)

	require.NoError(t, err)
	mockAthena.AssertExpectations(t)
	mockSNS.AssertExpectations(t) // Should be zero calls
}

func TestHandleEvent_QueryExecutionError(t *testing.T) {
	// Set required environment variables for test
	os.Setenv("ATHENA_DATABASE", "test-database")
	os.Setenv("ATHENA_RESULTS_BUCKET", "test-results-bucket")
	os.Setenv("ALERTS_SNS_TOPIC_ARN", "arn:aws:sns:ap-northeast-1:123456789012:test-alerts")
	defer func() {
		os.Unsetenv("ATHENA_DATABASE")
		os.Unsetenv("ATHENA_RESULTS_BUCKET")
		os.Unsetenv("ALERTS_SNS_TOPIC_ARN")
	}()

	mockAthena := new(MockAthenaAPI)
	mockSNS := new(MockSNSAPI)

	// Mock StartQueryExecution to return error
	mockAthena.On("StartQueryExecution", mock.Anything, mock.Anything).Return(
		(*athena.StartQueryExecutionOutput)(nil), assert.AnError).Once()

	handler := &Handler{
		athenaClient:   mockAthena,
		snsClient:      mockSNS,
		database:       "test-database",
		resultsBucket:  "test-results-bucket",
		alertsTopicArn: "arn:aws:sns:ap-northeast-1:123456789012:test-alerts",
		queries: []QueryDefinition{
			{Name: "test_query", Description: "Test", SQL: "SELECT 1"},
		},
	}

	cloudWatchEvent := events.CloudWatchEvent{
		Source: "aws.events",
		Detail: json.RawMessage(`{}`),
	}

	// Execute - should not return error, but should continue with other queries
	err := handler.HandleEvent(context.Background(), cloudWatchEvent)

	// Should not return error even if individual queries fail
	require.NoError(t, err)
	mockAthena.AssertExpectations(t)
}

func TestGetQueryResults_Processing(t *testing.T) {
	mockAthena := new(MockAthenaAPI)

	queryExecutionId := "test-execution-id"

	// Test with multiple result rows
	mockAthena.On("GetQueryResults", mock.Anything, &athena.GetQueryResultsInput{
		QueryExecutionId: aws.String(queryExecutionId),
		MaxResults:       aws.Int32(100),
	}).Return(&athena.GetQueryResultsOutput{
		ResultSet: &types.ResultSet{
			Rows: []types.Row{
				// Header row
				{
					Data: []types.Datum{
						{VarCharValue: aws.String("ip_address")},
						{VarCharValue: aws.String("failed_count")},
						{VarCharValue: aws.String("user")},
					},
				},
				// Data row 1
				{
					Data: []types.Datum{
						{VarCharValue: aws.String("192.168.1.100")},
						{VarCharValue: aws.String("5")},
						{VarCharValue: aws.String("alice")},
					},
				},
				// Data row 2
				{
					Data: []types.Datum{
						{VarCharValue: aws.String("192.168.1.101")},
						{VarCharValue: aws.String("3")},
						{VarCharValue: aws.String("bob")},
					},
				},
			},
		},
	}, nil)

	handler := &Handler{athenaClient: mockAthena}

	results, err := handler.getQueryResults(context.Background(), queryExecutionId)

	require.NoError(t, err)
	assert.Len(t, results, 2) // Should have 2 data rows

	// Verify first result
	assert.Equal(t, "192.168.1.100", results[0].Data["ip_address"])
	assert.Equal(t, "5", results[0].Data["failed_count"])
	assert.Equal(t, "alice", results[0].Data["user"])

	// Verify second result
	assert.Equal(t, "192.168.1.101", results[1].Data["ip_address"])
	assert.Equal(t, "3", results[1].Data["failed_count"])
	assert.Equal(t, "bob", results[1].Data["user"])

	mockAthena.AssertExpectations(t)
}

func TestNewHandler_MissingEnvironmentVariables(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func()
		cleanup func()
		errMsg  string
	}{
		{
			name: "missing ATHENA_DATABASE",
			setup: func() {
				os.Unsetenv("ATHENA_DATABASE")
				os.Setenv("ATHENA_RESULTS_BUCKET", "test-bucket")
				os.Setenv("ALERTS_SNS_TOPIC_ARN", "test-arn")
			},
			cleanup: func() {
				os.Unsetenv("ATHENA_RESULTS_BUCKET")
				os.Unsetenv("ALERTS_SNS_TOPIC_ARN")
			},
			errMsg: "ATHENA_DATABASE environment variable is required",
		},
		{
			name: "missing ATHENA_RESULTS_BUCKET",
			setup: func() {
				os.Setenv("ATHENA_DATABASE", "test-db")
				os.Unsetenv("ATHENA_RESULTS_BUCKET")
				os.Setenv("ALERTS_SNS_TOPIC_ARN", "test-arn")
			},
			cleanup: func() {
				os.Unsetenv("ATHENA_DATABASE")
				os.Unsetenv("ALERTS_SNS_TOPIC_ARN")
			},
			errMsg: "ATHENA_RESULTS_BUCKET environment variable is required",
		},
		{
			name: "missing ALERTS_SNS_TOPIC_ARN",
			setup: func() {
				os.Setenv("ATHENA_DATABASE", "test-db")
				os.Setenv("ATHENA_RESULTS_BUCKET", "test-bucket")
				os.Unsetenv("ALERTS_SNS_TOPIC_ARN")
			},
			cleanup: func() {
				os.Unsetenv("ATHENA_DATABASE")
				os.Unsetenv("ATHENA_RESULTS_BUCKET")
			},
			errMsg: "ALERTS_SNS_TOPIC_ARN environment variable is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()
			defer tc.cleanup()

			_, err := NewHandler()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.errMsg)
		})
	}
}
