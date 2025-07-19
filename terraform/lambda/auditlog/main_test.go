package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/m-mizutani/seccamp-2025-b1/internal/logcore"
)

func TestHandlerIdempotency(t *testing.T) {
	ctx := context.Background()
	// 現在日付を使用してテスト時刻を設定
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startTime := today.Add(10 * time.Hour).Format(time.RFC3339)
	endTime := today.Add(11 * time.Hour).Format(time.RFC3339)

	request := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		QueryStringParameters: map[string]string{
			"startTime": startTime,
			"endTime":   endTime,
		},
	}

	var responses []LogResponse
	const iterations = 10

	for i := 0; i < iterations; i++ {
		response, err := Handler(ctx, request)
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

	reference := responses[0]
	for i := 1; i < iterations; i++ {
		if !compareLogResponses(reference, responses[i]) {
			t.Errorf("Response %d differs from the first response", i)
			t.Logf("First response: %+v", reference)
			t.Logf("Response %d: %+v", i, responses[i])
		}
	}
}

func TestTimeShiftConsistency(t *testing.T) {
	ctx := context.Background()
	
	// 現在日付を使用
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	testCases := []struct {
		name           string
		startTime1     string
		endTime1       string
		startTime2     string
		endTime2       string
		expectedSubset bool
	}{
		{
			name:           "Overlapping time ranges",
			startTime1:     today.Add(10 * time.Hour).Format(time.RFC3339),
			endTime1:       today.Add(11 * time.Hour).Format(time.RFC3339),
			startTime2:     today.Add(10*time.Hour + 30*time.Minute).Format(time.RFC3339),
			endTime2:       today.Add(11*time.Hour + 30*time.Minute).Format(time.RFC3339),
			expectedSubset: true,
		},
		{
			name:           "Subset time range",
			startTime1:     today.Add(10 * time.Hour).Format(time.RFC3339),
			endTime1:       today.Add(12 * time.Hour).Format(time.RFC3339),
			startTime2:     today.Add(10*time.Hour + 30*time.Minute).Format(time.RFC3339),
			endTime2:       today.Add(11*time.Hour + 30*time.Minute).Format(time.RFC3339),
			expectedSubset: true,
		},
		{
			name:           "Adjacent time ranges",
			startTime1:     today.Add(10 * time.Hour).Format(time.RFC3339),
			endTime1:       today.Add(11 * time.Hour).Format(time.RFC3339),
			startTime2:     today.Add(11 * time.Hour).Format(time.RFC3339),
			endTime2:       today.Add(12 * time.Hour).Format(time.RFC3339),
			expectedSubset: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request1 := events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				QueryStringParameters: map[string]string{
					"startTime": tc.startTime1,
					"endTime":   tc.endTime1,
				},
			}

			request2 := events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				QueryStringParameters: map[string]string{
					"startTime": tc.startTime2,
					"endTime":   tc.endTime2,
				},
			}

			response1, err := Handler(ctx, request1)
			if err != nil {
				t.Fatalf("Handler returned error for request1: %v", err)
			}

			response2, err := Handler(ctx, request2)
			if err != nil {
				t.Fatalf("Handler returned error for request2: %v", err)
			}

			var logResp1, logResp2 LogResponse
			if err := json.Unmarshal([]byte(response1.Body), &logResp1); err != nil {
				t.Fatalf("Failed to unmarshal response1: %v", err)
			}
			if err := json.Unmarshal([]byte(response2.Body), &logResp2); err != nil {
				t.Fatalf("Failed to unmarshal response2: %v", err)
			}

			if tc.expectedSubset {
				if !hasTimeOverlap(tc.startTime1, tc.endTime1, tc.startTime2, tc.endTime2) {
					return
				}

				overlap := getOverlappingLogs(logResp1.Logs, logResp2.Logs, tc.startTime2, tc.endTime2)
				if len(overlap) == 0 && len(logResp2.Logs) > 0 {
					t.Errorf("Expected overlapping logs between time ranges, but found none")
				}
			}
		})
	}
}

func TestOffsetLimitIdempotency(t *testing.T) {
	ctx := context.Background()
	// 現在日付を使用
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startTime := today.Add(10 * time.Hour).Format(time.RFC3339)
	endTime := today.Add(11 * time.Hour).Format(time.RFC3339)

	baseRequest := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		QueryStringParameters: map[string]string{
			"startTime": startTime,
			"endTime":   endTime,
		},
	}

	baseResponse, err := Handler(ctx, baseRequest)
	if err != nil {
		t.Fatalf("Handler returned error for base request: %v", err)
	}

	var baseLogResp LogResponse
	if err := json.Unmarshal([]byte(baseResponse.Body), &baseLogResp); err != nil {
		t.Fatalf("Failed to unmarshal base response: %v", err)
	}

	totalLogs := baseLogResp.Metadata.Total

	testCases := []struct {
		name   string
		offset int
		limit  int
	}{
		{"First page", 0, 10},
		{"Second page", 10, 10},
		{"Partial page", 5, 15},
		{"Large limit", 0, 100},
		{"Mid-range", 20, 30},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				QueryStringParameters: map[string]string{
					"startTime": startTime,
					"endTime":   endTime,
					"offset":    fmt.Sprintf("%d", tc.offset),
					"limit":     fmt.Sprintf("%d", tc.limit),
				},
			}

			var responses []LogResponse
			for i := 0; i < 5; i++ {
				response, err := Handler(ctx, request)
				if err != nil {
					t.Fatalf("Handler returned error on iteration %d: %v", i, err)
				}

				var logResp LogResponse
				if err := json.Unmarshal([]byte(response.Body), &logResp); err != nil {
					t.Fatalf("Failed to unmarshal response on iteration %d: %v", i, err)
				}

				responses = append(responses, logResp)
			}

			for i := 1; i < len(responses); i++ {
				if !compareLogResponses(responses[0], responses[i]) {
					t.Errorf("Response %d differs from the first response for offset=%d, limit=%d", i, tc.offset, tc.limit)
				}
			}

			if responses[0].Metadata.Total != totalLogs {
				t.Errorf("Total count changed: expected %d, got %d", totalLogs, responses[0].Metadata.Total)
			}

			expectedCount := tc.limit
			if tc.offset+tc.limit > totalLogs {
				expectedCount = max(0, totalLogs-tc.offset)
			}
			if len(responses[0].Logs) != expectedCount {
				t.Errorf("Expected %d logs, got %d", expectedCount, len(responses[0].Logs))
			}
		})
	}
}

func TestPaginationConsistency(t *testing.T) {
	ctx := context.Background()
	// 現在日付を使用
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startTime := today.Add(10 * time.Hour).Format(time.RFC3339)
	endTime := today.Add(11 * time.Hour).Format(time.RFC3339)

	allLogsRequest := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		QueryStringParameters: map[string]string{
			"startTime": startTime,
			"endTime":   endTime,
			"limit":     "100",
		},
	}

	allLogsResponse, err := Handler(ctx, allLogsRequest)
	if err != nil {
		t.Fatalf("Handler returned error for all logs request: %v", err)
	}

	var allLogs LogResponse
	if err := json.Unmarshal([]byte(allLogsResponse.Body), &allLogs); err != nil {
		t.Fatalf("Failed to unmarshal all logs response: %v", err)
	}

	pageSize := 10
	var paginatedLogs []logcore.GoogleWorkspaceLogEntry

	for offset := 0; offset < allLogs.Metadata.Total; offset += pageSize {
		pageRequest := events.APIGatewayProxyRequest{
			HTTPMethod: "GET",
			QueryStringParameters: map[string]string{
				"startTime": startTime,
				"endTime":   endTime,
				"offset":    fmt.Sprintf("%d", offset),
				"limit":     fmt.Sprintf("%d", pageSize),
			},
		}

		pageResponse, err := Handler(ctx, pageRequest)
		if err != nil {
			t.Fatalf("Handler returned error for page at offset %d: %v", offset, err)
		}

		var pageData LogResponse
		if err := json.Unmarshal([]byte(pageResponse.Body), &pageData); err != nil {
			t.Fatalf("Failed to unmarshal page response at offset %d: %v", offset, err)
		}

		paginatedLogs = append(paginatedLogs, pageData.Logs...)
	}

	if len(paginatedLogs) != len(allLogs.Logs) {
		t.Errorf("Paginated logs count (%d) doesn't match all logs count (%d)", len(paginatedLogs), len(allLogs.Logs))
	}

	for i := 0; i < len(allLogs.Logs) && i < len(paginatedLogs); i++ {
		if !compareLogEntries(allLogs.Logs[i], paginatedLogs[i]) {
			t.Errorf("Log at index %d differs between all logs and paginated logs", i)
		}
	}
}

func compareLogResponses(a, b LogResponse) bool {
	if a.Date != b.Date {
		return false
	}

	if a.Metadata.Total != b.Metadata.Total ||
		a.Metadata.Offset != b.Metadata.Offset ||
		a.Metadata.Limit != b.Metadata.Limit {
		return false
	}

	if len(a.Logs) != len(b.Logs) {
		return false
	}

	for i := range a.Logs {
		if !compareLogEntries(a.Logs[i], b.Logs[i]) {
			return false
		}
	}

	return true
}

func compareLogEntries(a, b logcore.GoogleWorkspaceLogEntry) bool {
	aJSON, err1 := json.Marshal(a)
	bJSON, err2 := json.Marshal(b)

	if err1 != nil || err2 != nil {
		return false
	}

	return string(aJSON) == string(bJSON)
}

func hasTimeOverlap(start1Str, end1Str, start2Str, end2Str string) bool {
	start1, _ := time.Parse(time.RFC3339, start1Str)
	end1, _ := time.Parse(time.RFC3339, end1Str)
	start2, _ := time.Parse(time.RFC3339, start2Str)
	end2, _ := time.Parse(time.RFC3339, end2Str)

	return start1.Before(end2) && start2.Before(end1)
}

func getOverlappingLogs(logs1, logs2 []logcore.GoogleWorkspaceLogEntry, overlapStart, overlapEnd string) []logcore.GoogleWorkspaceLogEntry {
	start, _ := time.Parse(time.RFC3339, overlapStart)
	end, _ := time.Parse(time.RFC3339, overlapEnd)

	logMap := make(map[string]logcore.GoogleWorkspaceLogEntry)

	for _, log := range logs1 {
		logTime, err := time.Parse(time.RFC3339, log.ID.Time)
		if err == nil && !logTime.Before(start) && logTime.Before(end) {
			logMap[log.ID.UniqueQualifier] = log
		}
	}

	var overlapping []logcore.GoogleWorkspaceLogEntry
	for _, log := range logs2 {
		if _, exists := logMap[log.ID.UniqueQualifier]; exists {
			overlapping = append(overlapping, log)
		}
	}

	return overlapping
}

func TestFutureLogsFiltering(t *testing.T) {
	ctx := context.Background()
	// 現在時刻から1時間後まで（未来を含む）の範囲でリクエスト
	now := time.Now()
	startTime := now.Add(-30 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(30 * time.Minute).Format(time.RFC3339)

	request := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		QueryStringParameters: map[string]string{
			"startTime": startTime,
			"endTime":   endTime,
		},
	}

	response, err := Handler(ctx, request)
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
