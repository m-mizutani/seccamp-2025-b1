package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandlerConsistency(t *testing.T) {
	ctx := context.Background()
	// 現在日付を使用してテスト時刻を設定
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startTime := today.Add(10 * time.Hour).Format(time.RFC3339)
	endTime := today.Add(11 * time.Hour).Format(time.RFC3339)

	request := events.LambdaFunctionURLRequest{
		RequestContext: events.LambdaFunctionURLRequestContext{
			HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
				Method: "GET",
				Path:   "/",
			},
		},
		QueryStringParameters: map[string]string{
			"startTime": startTime,
			"endTime":   endTime,
		},
	}

	var responses []LogResponse
	const iterations = 5

	for i := 0; i < iterations; i++ {
		response, err := handler(ctx, request)
		if err != nil {
			t.Fatalf("Handler returned error on iteration %d: %v", i, err)
		}

		if response.StatusCode != 200 {
			t.Fatalf("Expected status 200, got %d on iteration %d", response.StatusCode, i)
		}

		var logResp LogResponse
		if err := json.Unmarshal([]byte(response.Body), &logResp); err != nil {
			t.Fatalf("Failed to unmarshal response on iteration %d: %v", i, err)
		}

		responses = append(responses, logResp)
	}

	// 基本的な一貫性チェック（完全な冪等性ではなく、主要な属性の一貫性）
	reference := responses[0]
	for i := 1; i < iterations; i++ {
		// 総数は同じであるべき
		if responses[i].Metadata.Total != reference.Metadata.Total {
			t.Errorf("Total count differs: iteration 0 has %d, iteration %d has %d",
				reference.Metadata.Total, i, responses[i].Metadata.Total)
		}

		// 返されるログ数は同じであるべき
		if len(responses[i].Logs) != len(reference.Logs) {
			t.Errorf("Log count differs: iteration 0 has %d logs, iteration %d has %d logs",
				len(reference.Logs), i, len(responses[i].Logs))
		}

		// 日付範囲は同じであるべき
		if responses[i].Date != reference.Date {
			t.Errorf("Date range differs: iteration 0 has %s, iteration %d has %s",
				reference.Date, i, responses[i].Date)
		}
	}

	t.Logf("All %d iterations returned consistent metadata: Total=%d, Count=%d",
		iterations, reference.Metadata.Total, len(reference.Logs))
}

func TestParameterValidation(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name           string
		startTime      string
		endTime        string
		limit          string
		offset         string
		expectedError  bool
		expectedStatus int
	}{
		{
			name:           "Valid parameters",
			startTime:      "2025-07-19T10:00:00Z",
			endTime:        "2025-07-19T11:00:00Z",
			limit:          "50",
			offset:         "0",
			expectedError:  false,
			expectedStatus: 200,
		},
		{
			name:           "Missing startTime",
			startTime:      "",
			endTime:        "2025-07-19T11:00:00Z",
			expectedError:  true,
			expectedStatus: 400,
		},
		{
			name:           "Invalid time format",
			startTime:      "invalid-time",
			endTime:        "2025-07-19T11:00:00Z",
			expectedError:  true,
			expectedStatus: 400,
		},
		{
			name:           "Limit exceeds maximum",
			startTime:      "2025-07-19T10:00:00Z",
			endTime:        "2025-07-19T11:00:00Z",
			limit:          "200",
			expectedError:  true,
			expectedStatus: 400,
		},
		{
			name:           "Negative offset",
			startTime:      "2025-07-19T10:00:00Z",
			endTime:        "2025-07-19T11:00:00Z",
			offset:         "-10",
			expectedError:  true,
			expectedStatus: 400,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := map[string]string{
				"startTime": tc.startTime,
				"endTime":   tc.endTime,
			}
			if tc.limit != "" {
				params["limit"] = tc.limit
			}
			if tc.offset != "" {
				params["offset"] = tc.offset
			}

			request := events.LambdaFunctionURLRequest{
				RequestContext: events.LambdaFunctionURLRequestContext{
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method: "GET",
						Path:   "/",
					},
				},
				QueryStringParameters: params,
			}

			response, err := handler(ctx, request)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if response.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, response.StatusCode)
			}
		})
	}
}

func TestFutureLogsFiltering(t *testing.T) {
	ctx := context.Background()
	// 現在時刻から1時間後まで（未来を含む）の範囲でリクエスト
	now := time.Now()
	startTime := now.Add(-30 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(30 * time.Minute).Format(time.RFC3339)

	request := events.LambdaFunctionURLRequest{
		RequestContext: events.LambdaFunctionURLRequestContext{
			HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
				Method: "GET",
				Path:   "/",
			},
		},
		QueryStringParameters: map[string]string{
			"startTime": startTime,
			"endTime":   endTime,
		},
	}

	response, err := handler(ctx, request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	var logResp LogResponse
	if err := json.Unmarshal([]byte(response.Body), &logResp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// すべてのログが現在時刻以前であることを確認
	for _, log := range logResp.Logs {
		logTime, err := time.Parse(time.RFC3339, log.ID.Time)
		if err != nil {
			t.Errorf("Failed to parse log time: %v", err)
			continue
		}

		if logTime.After(now) {
			t.Errorf("Found future log: %s (current time: %s)", log.ID.Time, now.Format(time.RFC3339))
		}
	}
}

func TestJSTTimeDistribution(t *testing.T) {
	ctx := context.Background()
	jst, _ := time.LoadLocation("Asia/Tokyo")

	// 各時間帯でテスト
	testCases := []struct {
		name        string
		hour        int
		expectedTop string // 最も多いはずのイベントタイプ
	}{
		{"深夜 (00:00-00:05)", 0, "login"},
		{"午前業務 (10:00-10:05)", 10, "access"},
		{"昼休み (12:00-12:05)", 12, "login"},
		{"午後業務 (15:00-15:05)", 15, "access"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 過去の日付でJST時刻を設定（未来のログを避けるため）
			yesterday := time.Now().AddDate(0, 0, -1).In(jst)
			baseTime := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), tc.hour, 0, 0, 0, jst)
			startTime := baseTime.Format(time.RFC3339)
			endTime := baseTime.Add(5 * time.Minute).Format(time.RFC3339)

			request := events.LambdaFunctionURLRequest{
				RequestContext: events.LambdaFunctionURLRequestContext{
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method: "GET",
						Path:   "/",
					},
				},
				QueryStringParameters: map[string]string{
					"startTime": startTime,
					"endTime":   endTime,
					"limit":     "100",
				},
			}

			response, err := handler(ctx, request)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			var logResp LogResponse
			if err := json.Unmarshal([]byte(response.Body), &logResp); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if len(logResp.Logs) == 0 {
				t.Skip("No logs generated for this time range")
			}

			// イベントタイプの分布を集計
			eventTypes := make(map[string]int)
			for _, log := range logResp.Logs {
				if len(log.Events) > 0 {
					eventTypes[log.Events[0].Type]++
				}
			}

			// 最頻出タイプを確認
			maxType := ""
			maxCount := 0
			for et, count := range eventTypes {
				if count > maxCount {
					maxCount = count
					maxType = et
				}
			}

			t.Logf("Hour %d:00 JST - Top event type: %s (%d/%d)", tc.hour, maxType, maxCount, len(logResp.Logs))

			// 期待される最頻出タイプが実際に最多かチェック（警告のみ）
			if maxType != tc.expectedTop {
				t.Logf("Expected %s to be dominant, but got %s", tc.expectedTop, maxType)
			}
		})
	}
}
