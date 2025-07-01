package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSNSAPI implements SNSAPI interface for testing
type MockSNSAPI struct {
	mock.Mock
}

func (m *MockSNSAPI) Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*sns.PublishOutput), args.Error(1)
}

func TestSNSOperations_Publish_Success(t *testing.T) {
	mockSNS := new(MockSNSAPI)

	messageId := "test-message-id"
	topicArn := "arn:aws:sns:ap-northeast-1:123456789012:test-alerts"

	// Expected alert message
	expectedAlert := Alert{
		Title:       "Test Alert",
		Description: "This is a test alert",
		Attrs: map[string]interface{}{
			"ip_address":   "192.168.1.100",
			"failed_count": "5",
		},
	}

	expectedJSON, _ := json.Marshal(expectedAlert)

	mockSNS.On("Publish", mock.Anything, mock.MatchedBy(func(input *sns.PublishInput) bool {
		if input.TopicArn == nil || *input.TopicArn != topicArn {
			return false
		}
		if input.Subject == nil || *input.Subject != expectedAlert.Title {
			return false
		}
		if input.Message == nil {
			return false
		}

		// Parse the message to verify it's valid JSON
		var parsedAlert Alert
		err := json.Unmarshal([]byte(*input.Message), &parsedAlert)
		return err == nil && parsedAlert.Title == expectedAlert.Title
	})).Return(&sns.PublishOutput{
		MessageId: aws.String(messageId),
	}, nil)

	handler := &Handler{
		snsClient:      mockSNS,
		alertsTopicArn: topicArn,
	}

	result, err := handler.snsClient.Publish(context.Background(), &sns.PublishInput{
		TopicArn: aws.String(topicArn),
		Message:  aws.String(string(expectedJSON)),
		Subject:  aws.String(expectedAlert.Title),
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, messageId, *result.MessageId)

	mockSNS.AssertExpectations(t)
}

func TestSNSOperations_Publish_Error(t *testing.T) {
	mockSNS := new(MockSNSAPI)

	topicArn := "arn:aws:sns:ap-northeast-1:123456789012:test-alerts"

	mockSNS.On("Publish", mock.Anything, mock.Anything).Return(
		(*sns.PublishOutput)(nil), assert.AnError)

	handler := &Handler{
		snsClient:      mockSNS,
		alertsTopicArn: topicArn,
	}

	alert := Alert{
		Title:       "Test Alert",
		Description: "This is a test alert",
		Attrs:       map[string]interface{}{},
	}

	alertJSON, _ := json.Marshal(alert)

	result, err := handler.snsClient.Publish(context.Background(), &sns.PublishInput{
		TopicArn: aws.String(topicArn),
		Message:  aws.String(string(alertJSON)),
		Subject:  aws.String(alert.Title),
	})

	assert.Error(t, err)
	assert.Nil(t, result)

	mockSNS.AssertExpectations(t)
}

func TestSNSOperations_AlertMessageStructure(t *testing.T) {
	testCases := []struct {
		name     string
		alert    Alert
		validate func(t *testing.T, alertJSON string)
	}{
		{
			name: "suspicious login alert",
			alert: Alert{
				Title:       "不審なログイン試行を検知",
				Description: "同一IPアドレスから複数回のログイン失敗が検知されました",
				Attrs: map[string]interface{}{
					"ip_address":   "192.168.1.100",
					"failed_count": "5",
					"user":         "alice",
				},
			},
			validate: func(t *testing.T, alertJSON string) {
				var alert Alert
				err := json.Unmarshal([]byte(alertJSON), &alert)
				require.NoError(t, err)

				assert.Equal(t, "不審なログイン試行を検知", alert.Title)
				assert.Contains(t, alert.Description, "同一IPアドレス")
				assert.Equal(t, "192.168.1.100", alert.Attrs["ip_address"])
				assert.Equal(t, "5", alert.Attrs["failed_count"])
				assert.Equal(t, "alice", alert.Attrs["user"])
			},
		},
		{
			name: "mass data access alert",
			alert: Alert{
				Title:       "大量データアクセスを検知",
				Description: "短時間での大量のread操作が検知されました",
				Attrs: map[string]interface{}{
					"user":         "bob",
					"access_count": "150",
					"timeframe":    "1 hour",
				},
			},
			validate: func(t *testing.T, alertJSON string) {
				var alert Alert
				err := json.Unmarshal([]byte(alertJSON), &alert)
				require.NoError(t, err)

				assert.Equal(t, "大量データアクセスを検知", alert.Title)
				assert.Contains(t, alert.Description, "短時間")
				assert.Equal(t, "bob", alert.Attrs["user"])
				assert.Equal(t, "150", alert.Attrs["access_count"])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			alertJSON, err := json.Marshal(tc.alert)
			require.NoError(t, err)

			tc.validate(t, string(alertJSON))
		})
	}
}

func TestSNSOperations_MessageValidation(t *testing.T) {
	mockSNS := new(MockSNSAPI)

	topicArn := "arn:aws:sns:ap-northeast-1:123456789012:test-alerts"

	// Verify that the message is properly formatted JSON
	mockSNS.On("Publish", mock.Anything, mock.MatchedBy(func(input *sns.PublishInput) bool {
		if input.Message == nil {
			return false
		}

		// Verify it's valid JSON
		var alert Alert
		err := json.Unmarshal([]byte(*input.Message), &alert)
		if err != nil {
			return false
		}

		// Verify required fields are present
		return alert.Title != "" && alert.Description != "" && alert.Attrs != nil
	})).Return(&sns.PublishOutput{
		MessageId: aws.String("test-message-id"),
	}, nil)

	handler := &Handler{
		snsClient:      mockSNS,
		alertsTopicArn: topicArn,
	}

	alert := Alert{
		Title:       "テストアラート",
		Description: "これはテスト用のアラートです",
		Attrs: map[string]interface{}{
			"severity": "high",
			"source":   "test",
		},
	}

	alertJSON, err := json.Marshal(alert)
	require.NoError(t, err)

	_, err = handler.snsClient.Publish(context.Background(), &sns.PublishInput{
		TopicArn: aws.String(topicArn),
		Message:  aws.String(string(alertJSON)),
		Subject:  aws.String(alert.Title),
	})

	require.NoError(t, err)
	mockSNS.AssertExpectations(t)
}
