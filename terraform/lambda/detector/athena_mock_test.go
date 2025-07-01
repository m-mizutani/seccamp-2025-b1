package main

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAthenaAPI implements AthenaAPI interface for testing
type MockAthenaAPI struct {
	mock.Mock
}

func (m *MockAthenaAPI) StartQueryExecution(ctx context.Context, params *athena.StartQueryExecutionInput, optFns ...func(*athena.Options)) (*athena.StartQueryExecutionOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*athena.StartQueryExecutionOutput), args.Error(1)
}

func (m *MockAthenaAPI) GetQueryExecution(ctx context.Context, params *athena.GetQueryExecutionInput, optFns ...func(*athena.Options)) (*athena.GetQueryExecutionOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*athena.GetQueryExecutionOutput), args.Error(1)
}

func (m *MockAthenaAPI) GetQueryResults(ctx context.Context, params *athena.GetQueryResultsInput, optFns ...func(*athena.Options)) (*athena.GetQueryResultsOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*athena.GetQueryResultsOutput), args.Error(1)
}

func TestAthenaOperations_StartQueryExecution_Success(t *testing.T) {
	mockAthena := new(MockAthenaAPI)

	queryExecutionId := "test-execution-id"

	mockAthena.On("StartQueryExecution", mock.Anything, mock.MatchedBy(func(input *athena.StartQueryExecutionInput) bool {
		return input.QueryString != nil && *input.QueryString != "" &&
			input.QueryExecutionContext != nil &&
			input.ResultConfiguration != nil
	})).Return(&athena.StartQueryExecutionOutput{
		QueryExecutionId: aws.String(queryExecutionId),
	}, nil)

	handler := &Handler{
		athenaClient:   mockAthena,
		database:       "test-database",
		resultsBucket:  "test-results-bucket",
		alertsTopicArn: "test-topic-arn",
	}

	result, err := handler.athenaClient.StartQueryExecution(context.Background(), &athena.StartQueryExecutionInput{
		QueryString: aws.String("SELECT * FROM test_table"),
		QueryExecutionContext: &types.QueryExecutionContext{
			Database: aws.String(handler.database),
		},
		ResultConfiguration: &types.ResultConfiguration{
			OutputLocation: aws.String("s3://test-results-bucket/"),
		},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, queryExecutionId, *result.QueryExecutionId)

	mockAthena.AssertExpectations(t)
}

func TestAthenaOperations_StartQueryExecution_Error(t *testing.T) {
	mockAthena := new(MockAthenaAPI)

	mockAthena.On("StartQueryExecution", mock.Anything, mock.Anything).Return(
		(*athena.StartQueryExecutionOutput)(nil), assert.AnError)

	handler := &Handler{
		athenaClient:   mockAthena,
		database:       "test-database",
		resultsBucket:  "test-results-bucket",
		alertsTopicArn: "test-topic-arn",
	}

	result, err := handler.athenaClient.StartQueryExecution(context.Background(), &athena.StartQueryExecutionInput{
		QueryString: aws.String("SELECT * FROM test_table"),
		QueryExecutionContext: &types.QueryExecutionContext{
			Database: aws.String(handler.database),
		},
		ResultConfiguration: &types.ResultConfiguration{
			OutputLocation: aws.String("s3://test-results-bucket/"),
		},
	})

	assert.Error(t, err)
	assert.Nil(t, result)

	mockAthena.AssertExpectations(t)
}

func TestAthenaOperations_GetQueryExecution_Success(t *testing.T) {
	mockAthena := new(MockAthenaAPI)

	queryExecutionId := "test-execution-id"

	mockAthena.On("GetQueryExecution", mock.Anything, &athena.GetQueryExecutionInput{
		QueryExecutionId: aws.String(queryExecutionId),
	}).Return(&athena.GetQueryExecutionOutput{
		QueryExecution: &types.QueryExecution{
			QueryExecutionId: aws.String(queryExecutionId),
			Status: &types.QueryExecutionStatus{
				State: types.QueryExecutionStateSucceeded,
			},
		},
	}, nil)

	handler := &Handler{athenaClient: mockAthena}

	result, err := handler.athenaClient.GetQueryExecution(context.Background(), &athena.GetQueryExecutionInput{
		QueryExecutionId: aws.String(queryExecutionId),
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, types.QueryExecutionStateSucceeded, result.QueryExecution.Status.State)

	mockAthena.AssertExpectations(t)
}

func TestAthenaOperations_GetQueryResults_WithData(t *testing.T) {
	mockAthena := new(MockAthenaAPI)

	queryExecutionId := "test-execution-id"

	// Mock response with header and data rows
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
				// Data row
				{
					Data: []types.Datum{
						{VarCharValue: aws.String("192.168.1.100")},
						{VarCharValue: aws.String("5")},
						{VarCharValue: aws.String("alice")},
					},
				},
			},
		},
	}, nil)

	handler := &Handler{athenaClient: mockAthena}

	result, err := handler.athenaClient.GetQueryResults(context.Background(), &athena.GetQueryResultsInput{
		QueryExecutionId: aws.String(queryExecutionId),
		MaxResults:       aws.Int32(100),
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.ResultSet.Rows, 2) // Header + 1 data row

	// Verify header
	assert.Equal(t, "ip_address", *result.ResultSet.Rows[0].Data[0].VarCharValue)
	assert.Equal(t, "failed_count", *result.ResultSet.Rows[0].Data[1].VarCharValue)
	assert.Equal(t, "user", *result.ResultSet.Rows[0].Data[2].VarCharValue)

	// Verify data
	assert.Equal(t, "192.168.1.100", *result.ResultSet.Rows[1].Data[0].VarCharValue)
	assert.Equal(t, "5", *result.ResultSet.Rows[1].Data[1].VarCharValue)
	assert.Equal(t, "alice", *result.ResultSet.Rows[1].Data[2].VarCharValue)

	mockAthena.AssertExpectations(t)
}

func TestAthenaOperations_GetQueryResults_EmptyData(t *testing.T) {
	mockAthena := new(MockAthenaAPI)

	queryExecutionId := "test-execution-id"

	// Mock response with only header row (no data)
	mockAthena.On("GetQueryResults", mock.Anything, &athena.GetQueryResultsInput{
		QueryExecutionId: aws.String(queryExecutionId),
		MaxResults:       aws.Int32(100),
	}).Return(&athena.GetQueryResultsOutput{
		ResultSet: &types.ResultSet{
			Rows: []types.Row{
				// Header row only
				{
					Data: []types.Datum{
						{VarCharValue: aws.String("ip_address")},
						{VarCharValue: aws.String("failed_count")},
					},
				},
			},
		},
	}, nil)

	handler := &Handler{athenaClient: mockAthena}

	result, err := handler.athenaClient.GetQueryResults(context.Background(), &athena.GetQueryResultsInput{
		QueryExecutionId: aws.String(queryExecutionId),
		MaxResults:       aws.Int32(100),
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.ResultSet.Rows, 1) // Only header row

	mockAthena.AssertExpectations(t)
}
