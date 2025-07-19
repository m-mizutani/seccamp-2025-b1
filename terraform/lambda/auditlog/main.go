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
//
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

	// JSTタイムゾーン
	jst, _ := time.LoadLocation("Asia/Tokyo")

	// リクエストされた日付を取得
	requestDate := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())

	// 指定時間範囲内のログを生成
	var allLogs []logcore.GoogleWorkspaceLogEntry

	// シードデータを複数回使用して、時間範囲全体をカバー
	duration := endTime.Sub(startTime)
	secondsInRange := int(duration.Seconds())

	// シードデータの全イベントタイプを収集
	seedsByType := make(map[uint8][]logcore.LogSeed)
	for _, seed := range dayTemplate.LogSeeds {
		seedsByType[seed.EventType] = append(seedsByType[seed.EventType], seed)
	}

	// 時間範囲内でログを生成（JST基準で分布を調整）
	logIndex := 0
	logInterval := 3 // 基本は3秒間隔

	for seconds := 0; seconds < secondsInRange; seconds += logInterval {
		logTime := startTime.Add(time.Duration(seconds) * time.Second)

		// 未来のログは除外
		if logTime.After(now) {
			break
		}

		// JST時刻を取得
		jstTime := logTime.In(jst)
		hour := jstTime.Hour()

		// 時間帯別のイベントタイプ分布を決定
		weights := getEventTypeWeights(hour)

		// 重み付けに基づいてイベントタイプを選択
		selectedType := selectEventTypeByWeight(weights, seconds)

		if seeds, ok := seedsByType[selectedType]; ok && len(seeds) > 0 {
			seedIndex := (seconds / logInterval) % len(seeds)
			seed := seeds[seedIndex]

			// 新しいシードを作成（タイムスタンプを調整）
			adjustedSeed := seed
			adjustedSeed.Timestamp = int64(seconds)

			logEntry := generator.GenerateLogEntry(adjustedSeed, requestDate, logIndex)
			logEntry.ID.Time = logTime.Format(time.RFC3339)
			allLogs = append(allLogs, *logEntry)
			logIndex++

			// ログインイベントは時間帯によって頻度を調整
			if selectedType == logcore.EventTypeLogin {
				extraLogs := getExtraLoginLogs(hour)
				for i := 0; i < extraLogs && seconds+i+1 < secondsInRange; i++ {
					extraTime := startTime.Add(time.Duration(seconds+i+1) * time.Second)
					if extraTime.After(now) {
						break
					}
					extraSeed := seed
					extraSeed.Timestamp = int64(seconds + i + 1)
					extraEntry := generator.GenerateLogEntry(extraSeed, requestDate, logIndex)
					extraEntry.ID.Time = extraTime.Format(time.RFC3339)
					allLogs = append(allLogs, *extraEntry)
					logIndex++
				}
			}
		}
	}

	total := len(allLogs)

	// ページネーション
	start := offset
	end := offset + limit
	if start >= total {
		return []logcore.GoogleWorkspaceLogEntry{}, total, nil
	}
	if end > total {
		end = total
	}

	return allLogs[start:end], total, nil
}

// JST時間帯別のイベントタイプ重み付け
func getEventTypeWeights(hour int) map[uint8]float64 {
	weights := make(map[uint8]float64)

	switch {
	case hour >= 0 && hour < 6: // 深夜〜早朝
		weights[logcore.EventTypeDriveAccess] = 0.05
		weights[logcore.EventTypeLogin] = 0.85 // ログインが多い（システムバッチ等）
		weights[logcore.EventTypeAdmin] = 0.05
		weights[logcore.EventTypeCalendar] = 0.05

	case hour >= 6 && hour < 9: // 朝（出勤時間）
		weights[logcore.EventTypeDriveAccess] = 0.15
		weights[logcore.EventTypeLogin] = 0.70 // 出勤時のログインが多い
		weights[logcore.EventTypeAdmin] = 0.05
		weights[logcore.EventTypeCalendar] = 0.10

	case hour >= 9 && hour < 12: // 午前（業務時間）
		weights[logcore.EventTypeDriveAccess] = 0.50 // 業務でDrive利用が活発
		weights[logcore.EventTypeLogin] = 0.15
		weights[logcore.EventTypeAdmin] = 0.15
		weights[logcore.EventTypeCalendar] = 0.20

	case hour >= 12 && hour < 13: // 昼休み
		weights[logcore.EventTypeDriveAccess] = 0.30
		weights[logcore.EventTypeLogin] = 0.40 // 昼休み後の再ログイン
		weights[logcore.EventTypeAdmin] = 0.10
		weights[logcore.EventTypeCalendar] = 0.20

	case hour >= 13 && hour < 18: // 午後（業務時間）
		weights[logcore.EventTypeDriveAccess] = 0.55 // 最も活発な時間帯
		weights[logcore.EventTypeLogin] = 0.10
		weights[logcore.EventTypeAdmin] = 0.15
		weights[logcore.EventTypeCalendar] = 0.20

	case hour >= 18 && hour < 21: // 夕方〜夜（残業時間）
		weights[logcore.EventTypeDriveAccess] = 0.35
		weights[logcore.EventTypeLogin] = 0.30 // 残業のための再ログイン
		weights[logcore.EventTypeAdmin] = 0.20 // 管理作業が増える
		weights[logcore.EventTypeCalendar] = 0.15

	default: // 21時以降
		weights[logcore.EventTypeDriveAccess] = 0.15
		weights[logcore.EventTypeLogin] = 0.60 // 深夜作業のログイン
		weights[logcore.EventTypeAdmin] = 0.20 // メンテナンス作業
		weights[logcore.EventTypeCalendar] = 0.05
	}

	return weights
}

// 重み付けに基づいてイベントタイプを選択
func selectEventTypeByWeight(weights map[uint8]float64, seed int) uint8 {
	// シード値から疑似乱数を生成
	rand := float64((seed*1103515245+12345)%100) / 100.0

	cumulative := 0.0
	for eventType, weight := range weights {
		cumulative += weight
		if rand < cumulative {
			return eventType
		}
	}

	// デフォルトはDriveAccess
	return logcore.EventTypeDriveAccess
}

// 時間帯に応じた追加ログイン数
func getExtraLoginLogs(hour int) int {
	switch {
	case hour >= 6 && hour < 9: // 朝の出勤時
		return 2
	case hour >= 12 && hour < 13: // 昼休み後
		return 1
	case hour >= 0 && hour < 6: // 深夜バッチ
		return 3
	default:
		return 0
	}
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
