package cmd

import (
	"context"
	"fmt"

	"github.com/m-mizutani/seccamp-2025-b1/internal/logcore"
	"github.com/urfave/cli/v3"
)

func StatsCommand() *cli.Command {
	return &cli.Command{
		Name:  "stats",
		Usage: "Show statistics of generated seeds",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "seeds",
				Usage:    "Path to seeds file",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return statsAction(c)
		},
	}
}

func statsAction(c *cli.Command) error {
	seedsPath := c.String("seeds")
	
	// シードファイル読み込み（自動判定）
	dayTemplate, err := loadDayTemplate(seedsPath)
	if err != nil {
		return fmt.Errorf("failed to load seeds file: %w", err)
	}
	
	// 統計情報計算
	stats := calculateStats(dayTemplate.LogSeeds)
	
	// 結果出力
	fmt.Printf("📊 Log Seeds Statistics\n")
	fmt.Printf("======================\n\n")
	
	fmt.Printf("📅 Date: %s\n", dayTemplate.Date)
	fmt.Printf("📈 Total Seeds: %d\n", stats.TotalSeeds)
	fmt.Printf("⏰ Generated: %s\n", dayTemplate.Metadata.Generated.Format("2006-01-02 15:04:05"))
	fmt.Printf("\n")
	
	fmt.Printf("🎯 Event Type Distribution:\n")
	for eventType, count := range stats.EventTypes {
		ratio := float64(count) / float64(stats.TotalSeeds) * 100
		fmt.Printf("  - Type %d: %d logs (%.1f%%)\n", eventType, count, ratio)
	}
	fmt.Printf("\n")
	
	fmt.Printf("⚠️ Anomaly Pattern Distribution:\n")
	normalCount := stats.Patterns[0]
	fmt.Printf("  - Normal: %d logs (%.1f%%)\n", normalCount, float64(normalCount)/float64(stats.TotalSeeds)*100)
	
	for pattern := uint8(1); pattern <= 10; pattern++ {
		if count, exists := stats.Patterns[pattern]; exists && count > 0 {
			ratio := float64(count) / float64(stats.TotalSeeds) * 100
			patternName := getPatternName(pattern)
			fmt.Printf("  - %s: %d logs (%.1f%%)\n", patternName, count, ratio)
		}
	}
	fmt.Printf("\n")
	
	fmt.Printf("👥 User Distribution:\n")
	for userIdx, count := range stats.Users {
		ratio := float64(count) / float64(stats.TotalSeeds) * 100
		fmt.Printf("  - User %d: %d logs (%.1f%%)\n", userIdx, count, ratio)
	}
	fmt.Printf("\n")
	
	fmt.Printf("📂 Resource Distribution:\n")
	for resourceIdx, count := range stats.Resources {
		ratio := float64(count) / float64(stats.TotalSeeds) * 100
		fmt.Printf("  - Resource %d: %d logs (%.1f%%)\n", resourceIdx, count, ratio)
	}
	fmt.Printf("\n")
	
	fmt.Printf("⏱️ Hourly Distribution:\n")
	for hour, count := range stats.HourlyDistribution {
		if count > 0 {
			ratio := float64(count) / float64(stats.TotalSeeds) * 100
			bar := generateBar(ratio, 50)
			fmt.Printf("  - %02d:00: %6d logs (%.1f%%) %s\n", hour, count, ratio, bar)
		}
	}
	
	return nil
}

type SeedStats struct {
	TotalSeeds          int
	EventTypes          map[uint8]int
	Patterns            map[uint8]int
	Users               map[uint8]int
	Resources           map[uint8]int
	HourlyDistribution  map[int]int
}

func calculateStats(seeds []logcore.LogSeed) SeedStats {
	stats := SeedStats{
		TotalSeeds:         len(seeds),
		EventTypes:         make(map[uint8]int),
		Patterns:           make(map[uint8]int),
		Users:              make(map[uint8]int),
		Resources:          make(map[uint8]int),
		HourlyDistribution: make(map[int]int),
	}
	
	for _, seed := range seeds {
		stats.EventTypes[seed.EventType]++
		stats.Patterns[seed.Pattern]++
		stats.Users[seed.UserIndex]++
		stats.Resources[seed.ResourceIdx]++
		
		// 時間別分布
		hour := int(seed.Timestamp / 3600)
		stats.HourlyDistribution[hour]++
	}
	
	return stats
}

func getPatternName(pattern uint8) string {
	switch pattern {
	case logcore.PatternExample1NightAdminDownload:
		return "Example1 Night Admin Download"
	case logcore.PatternExample2ExternalLinkAccess:
		return "Example2 External Link Access"
	case logcore.PatternExample3VpnLateralMovement:
		return "Example3 VPN Lateral Movement"
	case logcore.PatternTimeAnomaly:
		return "Time Anomaly"
	case logcore.PatternVolumeAnomaly:
		return "Volume Anomaly"
	default:
		return fmt.Sprintf("Pattern %d", pattern)
	}
}

func generateBar(ratio float64, maxWidth int) string {
	width := int(ratio / 100 * float64(maxWidth))
	if width < 0 {
		width = 0
	}
	if width > maxWidth {
		width = maxWidth
	}
	
	bar := ""
	for i := 0; i < width; i++ {
		bar += "█"
	}
	for i := width; i < maxWidth/4; i++ {
		bar += "░"
	}
	
	return bar
}