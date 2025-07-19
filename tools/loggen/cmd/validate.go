package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/m-mizutani/seccamp-2025-b1/internal/logcore"
	"github.com/urfave/cli/v3"
)

func ValidateCommand() *cli.Command {
	return &cli.Command{
		Name:  "validate",
		Usage: "Validate generated seeds file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "seeds",
				Usage:    "Path to seeds file",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Verbose output",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return validateAction(c)
		},
	}
}

func validateAction(c *cli.Command) error {
	seedsPath := c.String("seeds")
	verbose := c.Bool("verbose")

	// シードファイル読み込み（自動判定）
	dayTemplate, err := loadDayTemplate(seedsPath)
	if err != nil {
		return fmt.Errorf("failed to load seeds file: %w", err)
	}

	fmt.Printf("Validating seeds file: %s\n", seedsPath)
	fmt.Printf("Date: %s\n", dayTemplate.Date)
	fmt.Printf("Total seeds: %d\n", len(dayTemplate.LogSeeds))

	// バリデーション実行
	errors := []string{}

	// 1. 基本構造チェック
	if dayTemplate.Date == "" {
		errors = append(errors, "missing date field")
	}

	if len(dayTemplate.LogSeeds) == 0 {
		errors = append(errors, "no log seeds found")
	}

	// 2. 日付形式チェック
	if _, err := time.Parse("2006-01-02", dayTemplate.Date); err != nil {
		errors = append(errors, fmt.Sprintf("invalid date format: %v", err))
	}

	// 3. シード内容チェック
	anomalyCount := 0
	timeErrors := 0

	for i, seed := range dayTemplate.LogSeeds {
		// タイムスタンプ範囲チェック（0-86399秒）
		if seed.Timestamp < 0 || seed.Timestamp >= 86400 {
			if timeErrors < 5 {
				errors = append(errors, fmt.Sprintf("seed %d: invalid timestamp %d", i, seed.Timestamp))
			}
			timeErrors++
		}

		// パターンチェック
		if seed.Pattern > 0 {
			anomalyCount++
		}

		if seed.Pattern > 10 {
			errors = append(errors, fmt.Sprintf("seed %d: invalid pattern %d", i, seed.Pattern))
		}
	}

	if timeErrors > 5 {
		errors = append(errors, fmt.Sprintf("... and %d more timestamp errors", timeErrors-5))
	}

	// 4. 異常率チェック
	anomalyRatio := float64(anomalyCount) / float64(len(dayTemplate.LogSeeds))
	if anomalyRatio < 0.05 || anomalyRatio > 0.5 {
		errors = append(errors, fmt.Sprintf("unusual anomaly ratio: %.2f%% (%d/%d)",
			anomalyRatio*100, anomalyCount, len(dayTemplate.LogSeeds)))
	}

	// 5. バージョン互換性チェック
	if dayTemplate.Metadata.LogCoreVersion != logcore.LogCoreVersion {
		errors = append(errors, fmt.Sprintf("logcore version mismatch: file=%s, current=%s",
			dayTemplate.Metadata.LogCoreVersion, logcore.LogCoreVersion))
	}

	// 結果出力
	if len(errors) == 0 {
		fmt.Printf("✅ Validation passed!\n")
		if verbose {
			fmt.Printf("Details:\n")
			fmt.Printf("  - Anomaly logs: %d (%.2f%%)\n", anomalyCount, anomalyRatio*100)
			fmt.Printf("  - Generated: %s\n", dayTemplate.Metadata.Generated.Format(time.RFC3339))
			fmt.Printf("  - LogCore version: %s\n", dayTemplate.Metadata.LogCoreVersion)
		}
	} else {
		fmt.Printf("❌ Validation failed with %d errors:\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		return fmt.Errorf("validation failed")
	}

	return nil
}
