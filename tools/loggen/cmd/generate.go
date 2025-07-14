package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/m-mizutani/seccamp-2025-b1/internal/logcore"
	"github.com/m-mizutani/seccamp-2025-b1/tools/loggen/internal/seed"
	"github.com/urfave/cli/v3"
)

func GenerateCommand() *cli.Command {
	return &cli.Command{
		Name:  "generate",
		Usage: "Generate log seeds for a specific date",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "date",
				Usage: "Target date (YYYY-MM-DD)",
				Value: time.Now().Format("2006-01-02"),
			},
			&cli.StringFlag{
				Name:  "output",
				Usage: "Output directory",
				Value: "./output",
			},
			&cli.Float64Flag{
				Name:  "anomaly-ratio",
				Usage: "Anomaly log ratio (0.0-1.0)",
				Value: 0.15,
			},
			&cli.StringFlag{
				Name:  "format",
				Usage: "Output format (json, binary, binary-compressed)",
				Value: "binary-compressed",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Show what would be generated without writing files",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return generateAction(c)
		},
	}
}

func generateAction(c *cli.Command) error {
	dateStr := c.String("date")
	outputDir := c.String("output")
	anomalyRatio := c.Float64("anomaly-ratio")
	format := c.String("format")
	dryRun := c.Bool("dry-run")
	
	// 日付パース
	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}
	
	// シード生成
	fmt.Printf("Generating log seeds for %s...\n", targetDate.Format("2006-01-02"))
	
	generator := seed.NewGenerator()
	dayTemplate, err := generator.GenerateDayTemplate(targetDate, anomalyRatio)
	if err != nil {
		return fmt.Errorf("failed to generate day template: %w", err)
	}
	
	fmt.Printf("Generated %d log seeds\n", len(dayTemplate.LogSeeds))
	
	// 異常パターンの統計を計算
	totalAnomalies := 0
	for _, count := range dayTemplate.Metadata.AnomalyStats {
		totalAnomalies += count
	}
	
	fmt.Printf("Normal logs: %d\n", dayTemplate.Metadata.TotalLogs-totalAnomalies)
	fmt.Printf("Anomaly logs: %d\n", totalAnomalies)
	fmt.Printf("Anomaly ratio: %.2f%%\n", float64(totalAnomalies)/float64(dayTemplate.Metadata.TotalLogs)*100)
	
	if dryRun {
		fmt.Printf("Dry run mode - no files written (format would be: %s)\n", format)
		return nil
	}
	
	// 出力ディレクトリ作成
	seedsDir := filepath.Join(outputDir, "seeds")
	if err := os.MkdirAll(seedsDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// フォーマットに応じてファイル保存
	return saveSeeds(dayTemplate, seedsDir, targetDate, format)
}

func saveSeeds(dayTemplate *logcore.DayTemplate, seedsDir string, targetDate time.Time, format string) error {
	dateStr := targetDate.Format("2006-01-02")
	
	switch format {
	case "json":
		return saveAsJSON(dayTemplate, seedsDir, dateStr)
	case "binary":
		return saveAsBinary(dayTemplate, seedsDir, dateStr)
	case "binary-compressed":
		return saveAsBinaryCompressed(dayTemplate, seedsDir, dateStr)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func saveAsJSON(dayTemplate *logcore.DayTemplate, seedsDir, dateStr string) error {
	outputPath := filepath.Join(seedsDir, fmt.Sprintf("day_%s.json", dateStr))
	
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(dayTemplate); err != nil {
		return fmt.Errorf("failed to encode day template: %w", err)
	}
	
	// ファイルサイズ表示
	fileInfo, _ := file.Stat()
	fmt.Printf("Seeds written to: %s (size: %.2f MB)\n", outputPath, float64(fileInfo.Size())/1024/1024)
	
	return nil
}

func saveAsBinary(dayTemplate *logcore.DayTemplate, seedsDir, dateStr string) error {
	outputPath := filepath.Join(seedsDir, fmt.Sprintf("day_%s.bin", dateStr))
	
	data, err := dayTemplate.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal binary: %w", err)
	}
	
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write binary file: %w", err)
	}
	
	fmt.Printf("Seeds written to: %s (size: %.2f MB)\n", outputPath, float64(len(data))/1024/1024)
	
	return nil
}

func saveAsBinaryCompressed(dayTemplate *logcore.DayTemplate, seedsDir, dateStr string) error {
	outputPath := filepath.Join(seedsDir, fmt.Sprintf("day_%s.bin.gz", dateStr))
	
	data, err := dayTemplate.MarshalBinaryCompressed()
	if err != nil {
		return fmt.Errorf("failed to marshal binary compressed: %w", err)
	}
	
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write compressed file: %w", err)
	}
	
	fmt.Printf("Seeds written to: %s (size: %.2f MB)\n", outputPath, float64(len(data))/1024/1024)
	
	return nil
}