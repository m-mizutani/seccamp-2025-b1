package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
)

func CompareCommand() *cli.Command {
	return &cli.Command{
		Name:  "compare",
		Usage: "Compare file sizes of different formats",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "seeds-dir",
				Usage: "Directory containing seed files",
				Value: "./output/seeds",
			},
			&cli.StringFlag{
				Name:     "date",
				Usage:    "Date to compare (YYYY-MM-DD)",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return compareAction(c)
		},
	}
}

func compareAction(c *cli.Command) error {
	seedsDir := c.String("seeds-dir")
	date := c.String("date")

	// ファイルパス生成
	jsonFile := filepath.Join(seedsDir, fmt.Sprintf("day_%s.json", date))
	binaryFile := filepath.Join(seedsDir, fmt.Sprintf("day_%s.bin", date))
	compressedFile := filepath.Join(seedsDir, fmt.Sprintf("day_%s.bin.gz", date))

	fmt.Printf("📊 File Size Comparison for %s\n", date)
	fmt.Printf("=====================================\n\n")

	var jsonSize, binarySize, compressedSize int64

	// JSON形式
	if info, err := os.Stat(jsonFile); err == nil {
		jsonSize = info.Size()
		fmt.Printf("📄 JSON Format:        %s (%s)\n", formatFileSize(jsonSize), jsonFile)
	} else {
		fmt.Printf("📄 JSON Format:        File not found\n")
	}

	// バイナリ形式
	if info, err := os.Stat(binaryFile); err == nil {
		binarySize = info.Size()
		fmt.Printf("🔧 Binary Format:      %s (%s)\n", formatFileSize(binarySize), binaryFile)
		if jsonSize > 0 {
			reduction := calculateCompressionRatio(jsonSize, binarySize)
			fmt.Printf("   └── vs JSON:        %.1f%% reduction\n", reduction)
		}
	} else {
		fmt.Printf("🔧 Binary Format:      File not found\n")
	}

	// 圧縮バイナリ形式
	if info, err := os.Stat(compressedFile); err == nil {
		compressedSize = info.Size()
		fmt.Printf("🗜️  Compressed Binary: %s (%s)\n", formatFileSize(compressedSize), compressedFile)
		if jsonSize > 0 {
			reduction := calculateCompressionRatio(jsonSize, compressedSize)
			fmt.Printf("   └── vs JSON:        %.1f%% reduction\n", reduction)
		}
		if binarySize > 0 {
			reduction := calculateCompressionRatio(binarySize, compressedSize)
			fmt.Printf("   └── vs Binary:      %.1f%% reduction\n", reduction)
		}
	} else {
		fmt.Printf("🗜️  Compressed Binary: File not found\n")
	}

	fmt.Printf("\n")

	// 総評
	if jsonSize > 0 && compressedSize > 0 {
		totalReduction := calculateCompressionRatio(jsonSize, compressedSize)
		fmt.Printf("🎯 Best Performance: Compressed Binary\n")
		fmt.Printf("   └── Overall savings: %.1f%% (%.1fx smaller)\n",
			totalReduction, float64(jsonSize)/float64(compressedSize))
	}

	// 性能詳細
	fmt.Printf("\n📈 Performance Analysis:\n")
	fmt.Printf("  • JSON:        Human-readable, largest size, slower I/O\n")
	fmt.Printf("  • Binary:      Fast I/O, ~90%% size reduction\n")
	fmt.Printf("  • Compressed:  Smallest size, ~94%% reduction, good I/O\n")

	return nil
}
