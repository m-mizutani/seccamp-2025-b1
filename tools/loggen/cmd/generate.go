package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
				Usage: "Output destination (local directory or s3://bucket/prefix/)",
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
			&cli.IntFlag{
				Name:  "multiplier",
				Usage: "Multiply seed data by this factor",
				Value: 1,
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Show what would be generated without writing files",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return generateAction(ctx, c)
		},
	}
}

func generateAction(ctx context.Context, c *cli.Command) error {
	dateStr := c.String("date")
	output := c.String("output")
	anomalyRatio := c.Float64("anomaly-ratio")
	format := c.String("format")
	multiplier := c.Int("multiplier")
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

	// Multiply seeds if requested
	if multiplier > 1 {
		dayTemplate = multiplySeeds(dayTemplate, multiplier)
		fmt.Printf("Multiplied seeds by %dx\n", multiplier)
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

	// Check if output is S3
	if strings.HasPrefix(output, "s3://") {
		return saveToS3(ctx, dayTemplate, output, targetDate, format)
	}

	// Local file output
	seedsDir := filepath.Join(output, "seeds")
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

// multiplySeeds multiplies the seed data by the given factor
func multiplySeeds(template *logcore.DayTemplate, multiplier int) *logcore.DayTemplate {
	newTemplate := &logcore.DayTemplate{
		Date:     template.Date,
		LogSeeds: make([]logcore.LogSeed, 0, len(template.LogSeeds)*multiplier),
		Metadata: template.Metadata,
	}

	for i := 0; i < multiplier; i++ {
		for j, seed := range template.LogSeeds {
			adjustedSeed := seed
			// Spread timestamps across the day
			adjustedSeed.Timestamp = seed.Timestamp + int64(i*8640)
			// Vary user and resource indices
			if i%2 == 0 {
				adjustedSeed.UserIndex = uint8((int(seed.UserIndex) + i) % 256)
			}
			if i%3 == 0 {
				adjustedSeed.ResourceIdx = uint8((int(seed.ResourceIdx) + i) % 256)
			}
			// Vary the seed value for different random generation
			adjustedSeed.Seed = seed.Seed + uint32(i*1000+j)
			
			newTemplate.LogSeeds = append(newTemplate.LogSeeds, adjustedSeed)
		}
	}

	// Update metadata
	newTemplate.Metadata.TotalLogs = len(newTemplate.LogSeeds)
	
	return newTemplate
}

// saveToS3 saves the seed data to S3
func saveToS3(ctx context.Context, dayTemplate *logcore.DayTemplate, s3Path string, targetDate time.Time, format string) error {
	// Parse S3 path
	s3Path = strings.TrimPrefix(s3Path, "s3://")
	parts := strings.SplitN(s3Path, "/", 2)
	if len(parts) < 1 {
		return fmt.Errorf("invalid S3 path")
	}
	
	bucket := parts[0]
	prefix := ""
	if len(parts) > 1 {
		prefix = parts[1]
	}

	// Generate filename
	dateStr := targetDate.Format("2006-01-02")
	var filename string
	var data []byte
	var err error

	switch format {
	case "json":
		filename = fmt.Sprintf("day_%s.json", dateStr)
		var buf bytes.Buffer
		encoder := json.NewEncoder(&buf)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(dayTemplate); err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
		data = buf.Bytes()
	case "binary":
		filename = fmt.Sprintf("day_%s.bin", dateStr)
		data, err = dayTemplate.MarshalBinary()
		if err != nil {
			return fmt.Errorf("failed to marshal binary: %w", err)
		}
	case "binary-compressed":
		filename = fmt.Sprintf("large-seed.bin.gz")  // Fixed name for Lambda
		data, err = dayTemplate.MarshalBinaryCompressed()
		if err != nil {
			return fmt.Errorf("failed to marshal binary compressed: %w", err)
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Load AWS config
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}
	if region == "" {
		region = "ap-northeast-1" // デフォルトリージョン
	}
	
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg)

	// Upload to S3
	key := filepath.Join(prefix, "seeds", filename)
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	fmt.Printf("Seeds uploaded to: s3://%s/%s (size: %.2f MB)\n", bucket, key, float64(len(data))/1024/1024)
	
	return nil
}
