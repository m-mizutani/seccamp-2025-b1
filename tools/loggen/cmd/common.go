package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/m-mizutani/seccamp-2025-b1/internal/logcore"
)

// ファイル形式を自動判定してDayTemplateを読み込み
func loadDayTemplate(filePath string) (*logcore.DayTemplate, error) {
	// ファイル拡張子で形式を判定
	ext := strings.ToLower(filepath.Ext(filePath))

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var dayTemplate logcore.DayTemplate

	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &dayTemplate); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
	case ".bin":
		if err := dayTemplate.UnmarshalBinary(data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal binary: %w", err)
		}
	case ".gz":
		if err := dayTemplate.UnmarshalBinaryCompressed(data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal compressed binary: %w", err)
		}
	default:
		// 拡張子不明の場合、ファイル内容で判定
		if err := tryLoadAnyFormat(&dayTemplate, data); err != nil {
			return nil, fmt.Errorf("failed to load file (unknown format): %w", err)
		}
	}

	return &dayTemplate, nil
}

// 複数の形式を試して読み込み
func tryLoadAnyFormat(dayTemplate *logcore.DayTemplate, data []byte) error {
	// 1. JSON形式を試行
	if err := json.Unmarshal(data, dayTemplate); err == nil {
		return nil
	}

	// 2. バイナリ形式を試行
	if err := dayTemplate.UnmarshalBinary(data); err == nil {
		return nil
	}

	// 3. 圧縮バイナリ形式を試行
	if err := dayTemplate.UnmarshalBinaryCompressed(data); err == nil {
		return nil
	}

	return fmt.Errorf("unsupported file format")
}

// ファイルサイズを人間が読みやすい形式で表示
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// 圧縮率を計算
func calculateCompressionRatio(originalSize, compressedSize int64) float64 {
	if originalSize == 0 {
		return 0
	}
	return (1.0 - float64(compressedSize)/float64(originalSize)) * 100
}
