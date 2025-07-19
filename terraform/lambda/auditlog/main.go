package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/seccamp-2025-b1/internal/logcore"
)

// 埋め込まれたシードファイル（バイナリ圧縮形式）
//go:embed seeds/day_2024-08-12.bin.gz
var seedData []byte

// Lambda関数のレスポンス構造
type LogResponse struct {
	Date     string                            `json:"date"`
	Metadata ResponseMetadata                  `json:"metadata"`
	Logs     []logcore.GoogleWorkspaceLogEntry `json:"logs"`
}

type ResponseMetadata struct {
	Total     int       `json:"total"`
	Offset    int       `json:"offset"`
	Limit     int       `json:"limit"`
	Generated time.Time `json:"generated"`
}

// エラーレスポンス構造
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// CORS ヘッダー
	headers := map[string]string{
		"Content-Type":                 "application/json",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type",
	}

	// OPTIONS リクエスト（CORS プリフライト）
	if request.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    headers,
		}, nil
	}

	// クエリパラメータの解析
	var startTime, endTime time.Time
	var err error

	// startTime (必須)
	startTimeStr := request.QueryStringParameters["startTime"]
	if startTimeStr == "" {
		return errorResponse(400, "missing required parameter: startTime", headers)
	}
	startTime, err = time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return errorResponse(400, "invalid startTime format, use RFC3339 (2006-01-02T15:04:05Z)", headers)
	}

	// endTime (必須)
	endTimeStr := request.QueryStringParameters["endTime"]
	if endTimeStr == "" {
		return errorResponse(400, "missing required parameter: endTime", headers)
	}
	endTime, err = time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		return errorResponse(400, "invalid endTime format, use RFC3339 (2006-01-02T15:04:05Z)", headers)
	}

	// 時刻範囲検証
	if !endTime.After(startTime) {
		return errorResponse(400, "endTime must be after startTime", headers)
	}

	// limit (オプション、デフォルト100、最大100)
	limit := 100
	if limitStr := request.QueryStringParameters["limit"]; limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			if l <= 0 {
				return errorResponse(400, "limit must be greater than 0", headers)
			}
			if l > 100 {
				return errorResponse(400, "limit must not exceed 100", headers)
			}
			limit = l
		} else {
			return errorResponse(400, "invalid limit parameter", headers)
		}
	}

	// offset (オプション、デフォルト0)
	offset := 0
	if offsetStr := request.QueryStringParameters["offset"]; offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		} else {
			return errorResponse(400, "invalid offset parameter", headers)
		}
	}

	// ログ生成
	logs, total, err := generateLogs(startTime, endTime, limit, offset)
	if err != nil {
		return errorResponse(500, fmt.Sprintf("failed to generate logs: %v", err), headers)
	}

	response := LogResponse{
		Date: fmt.Sprintf("%s to %s", startTime.Format("2006-01-02T15:04:05Z"), endTime.Format("2006-01-02T15:04:05Z")),
		Metadata: ResponseMetadata{
			Total:     total,
			Offset:    offset,
			Limit:     limit,
			Generated: time.Now(),
		},
		Logs: logs,
	}

	body, err := json.Marshal(response)
	if err != nil {
		return errorResponse(500, "failed to marshal response", headers)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       string(body),
	}, nil
}

func generateLogs(startTime, endTime time.Time, limit, offset int) ([]logcore.GoogleWorkspaceLogEntry, int, error) {
	// 埋め込みシードデータの読み込み
	var dayTemplate logcore.DayTemplate
	if err := dayTemplate.UnmarshalBinaryCompressed(seedData); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal seed data: %w", err)
	}

	// 設定読み込み
	config := logcore.DefaultConfig()
	generator := logcore.NewGenerator(config)

	// 現在時刻を取得
	now := time.Now()
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 指定時間範囲内のログをフィルタリング
	var filteredLogs []logcore.GoogleWorkspaceLogEntry
	for i, seed := range dayTemplate.LogSeeds {
		// シードの時刻を現在日付に調整（時刻部分はseedのタイムスタンプを使用）
		logTime := currentDate.Add(time.Duration(seed.Timestamp) * time.Second)
		
		// 未来のログは除外
		if logTime.After(now) {
			continue
		}
		
		// 指定時間範囲内かチェック
		if (logTime.Equal(startTime) || logTime.After(startTime)) && logTime.Before(endTime) {
			logEntry := generator.GenerateLogEntry(seed, currentDate, i)
			// ログエントリの時刻を調整
			logEntry.ID.Time = logTime.Format(time.RFC3339)
			filteredLogs = append(filteredLogs, *logEntry)
		}
	}

	total := len(filteredLogs)
	
	// ページネーション
	start := offset
	end := offset + limit
	if start >= total {
		return []logcore.GoogleWorkspaceLogEntry{}, total, nil
	}
	if end > total {
		end = total
	}

	return filteredLogs[start:end], total, nil
}

func errorResponse(statusCode int, message string, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	errorResp := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}
	
	body, _ := json.Marshal(errorResp)
	
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       string(body),
	}, nil
}