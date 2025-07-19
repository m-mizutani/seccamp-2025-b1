package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/m-mizutani/seccamp-2025-b1/internal/logcore"
	"github.com/urfave/cli/v3"
)

func PreviewCommand() *cli.Command {
	return &cli.Command{
		Name:  "preview",
		Usage: "Preview logs generated from seeds",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "seeds",
				Usage:    "Path to seeds file",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "time-range",
				Usage: "Time range (e.g., '10:00-11:00')",
				Value: "10:00-11:00",
			},
			&cli.IntFlag{
				Name:  "limit",
				Usage: "Number of logs to preview",
				Value: 10,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return previewAction(c)
		},
	}
}

func previewAction(c *cli.Command) error {
	seedsPath := c.String("seeds")
	timeRange := c.String("time-range")
	limit := c.Int("limit")

	// シードファイル読み込み（自動判定）
	dayTemplate, err := loadDayTemplate(seedsPath)
	if err != nil {
		return fmt.Errorf("failed to load seeds file: %w", err)
	}

	// 時間範囲をパース
	startTime, endTime, err := parseTimeRange(timeRange, dayTemplate.Date)
	if err != nil {
		return fmt.Errorf("failed to parse time range: %w", err)
	}

	// 範囲内のシードを抽出
	seeds := logcore.ExtractSeedsInRange(dayTemplate, startTime, endTime)

	if len(seeds) == 0 {
		fmt.Printf("No logs found in time range %s\n", timeRange)
		return nil
	}

	fmt.Printf("Found %d logs in time range %s\n", len(seeds), timeRange)
	fmt.Printf("Showing first %d logs:\n\n", min(limit, len(seeds)))

	// ログ生成とプレビュー
	config := logcore.DefaultConfig()
	generator := logcore.NewGenerator(config)

	baseDate, _ := time.Parse("2006-01-02", dayTemplate.Date)

	for i, seed := range seeds[:min(limit, len(seeds))] {
		logEntry := generator.GenerateLogEntry(seed, baseDate, i)

		// JSON形式で出力
		jsonBytes, _ := json.MarshalIndent(logEntry, "", "  ")
		fmt.Printf("Log %d:\n%s\n\n", i+1, string(jsonBytes))
	}

	return nil
}

func parseTimeRange(timeRangeStr, dateStr string) (time.Time, time.Time, error) {
	parts := strings.Split(timeRangeStr, "-")
	if len(parts) != 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid time range format")
	}

	startTime, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", dateStr, parts[0]))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	endTime, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", dateStr, parts[1]))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return startTime, endTime, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
