package logcore

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// 異常パターンジェネレータ
type AnomalyGenerator struct {
	Patterns map[uint8]AnomalyFunc
}

type AnomalyFunc func(*GoogleWorkspaceLogEntry, *rand.Rand) *GoogleWorkspaceLogEntry

// 新しい異常パターンジェネレータを作成
func NewAnomalyGenerator() *AnomalyGenerator {
	return &AnomalyGenerator{
		Patterns: map[uint8]AnomalyFunc{
			PatternExample1NightAdminDownload: generateExample1NightAdminDownload,
			PatternExample2ExternalLinkAccess: generateExample2ExternalLinkAccess,
			PatternExample3VpnLateralMovement: generateExample3VpnLateralMovement,
			PatternTimeAnomaly:                generateTimeAnomaly,
			PatternVolumeAnomaly:              generateVolumeAnomaly,
		},
	}
}

// 異常パターンを適用
func (ag *AnomalyGenerator) ApplyAnomalyPattern(base *GoogleWorkspaceLogEntry, pattern uint8, rng *rand.Rand) *GoogleWorkspaceLogEntry {
	if anomalyFunc, exists := ag.Patterns[pattern]; exists {
		return anomalyFunc(base, rng)
	}
	return base
}

// 実例1: 夜間の管理者による大量学習データダウンロード
func generateExample1NightAdminDownload(base *GoogleWorkspaceLogEntry, rng *rand.Rand) *GoogleWorkspaceLogEntry {
	if strings.Contains(base.Actor.Email, "admin") || strings.Contains(base.Actor.Email, "manager") {
		timestamp, _ := time.Parse(time.RFC3339Nano, base.ID.Time)
		hour := timestamp.Hour()

		// 業務時間外（18:00-9:00）での異常動作
		if hour >= 18 || hour <= 9 {
			base.ID.ApplicationName = "drive"
			base.Events = []Event{
				{
					Type: "access",
					Name: "download",
					Parameters: []Parameter{
						{Name: "doc_id", Value: generateDocumentID(rng)},
						{Name: "doc_title", Value: "学習進捗データ_" + generateRandomDataset(rng)},
						{Name: "doc_type", Value: "spreadsheet"},
						{Name: "owner", Value: base.Actor.Email},
						{Name: "visibility", Value: "private"},
						{Name: "billable", BoolValue: true},
						{Name: "primary_event", BoolValue: true},
					},
				},
			}

			// 内部IPアドレス
			base.IPAddress = generateInternalIP(rng)
		}
	}
	return base
}

// 実例2: anyone with link設定ミスによる機密情報の意図しない外部流出
func generateExample2ExternalLinkAccess(base *GoogleWorkspaceLogEntry, rng *rand.Rand) *GoogleWorkspaceLogEntry {
	// 外部IPからのアクセス
	base.IPAddress = generateExternalIP(rng)
	base.Actor.Email = "unknown@external-domain.com"
	base.Actor.ProfileID = generateExternalProfileID(rng)

	// 機密フォルダへのアクセス
	confidentialFiles := []string{
		"学籍管理データベース.xlsx",
		"教職員人事データ.xlsx",
		"財務予算計画.xlsx",
		"試験問題集.docx",
	}
	selectedFile := confidentialFiles[rng.Intn(len(confidentialFiles))]

	base.ID.ApplicationName = "drive"
	base.Events = []Event{
		{
			Type: "access",
			Name: "view",
			Parameters: []Parameter{
				{Name: "doc_id", Value: generateDocumentID(rng)},
				{Name: "doc_title", Value: selectedFile},
				{Name: "doc_type", Value: "spreadsheet"},
				{Name: "owner", Value: "staff@muhai-academy.com"},
				{Name: "visibility", Value: "anyone_with_link"}, // 問題のある権限設定
				{Name: "primary_event", BoolValue: true},
			},
		},
	}

	return base
}

// 実例3: VPN脆弱性経由の攻撃 - Google Workspaceでの異常アクセス試行
func generateExample3VpnLateralMovement(base *GoogleWorkspaceLogEntry, rng *rand.Rand) *GoogleWorkspaceLogEntry {
	timestamp, _ := time.Parse(time.RFC3339Nano, base.ID.Time)
	hour := timestamp.Hour()

	// 業務時間内だが怪しいアクセスパターン
	if hour >= 9 && hour <= 18 {
		base.Actor.Email = "compromised.user@muhai-academy.com"

		// Google Workspace内のファイル・フォルダへの連続アクセス拒否
		workspaceTargets := []string{
			"財務データ/予算計画.xlsx",
			"学生記録/成績データ.xlsx",
			"教職員データ/人事情報.xlsx",
			"システム/バックアップ.zip",
		}
		selectedTarget := workspaceTargets[rng.Intn(len(workspaceTargets))]

		base.ID.ApplicationName = "drive"
		base.Events = []Event{
			{
				Type: "access",
				Name: "access_denied",
				Parameters: []Parameter{
					{Name: "doc_id", Value: generateDocumentID(rng)},
					{Name: "doc_title", Value: selectedTarget},
					{Name: "doc_type", Value: "spreadsheet"},
					{Name: "owner", Value: "admin@muhai-academy.com"},
					{Name: "visibility", Value: "private"},
					{Name: "denied_reason", Value: "insufficient_permissions"},
					{Name: "primary_event", BoolValue: true},
				},
			},
		}

		// VPN経由の内部IPからのアクセス
		base.IPAddress = generateVPNInternalIP(rng)
	}

	return base
}

// 時間外アクセス異常
func generateTimeAnomaly(base *GoogleWorkspaceLogEntry, rng *rand.Rand) *GoogleWorkspaceLogEntry {
	timestamp, _ := time.Parse(time.RFC3339Nano, base.ID.Time)
	hour := timestamp.Hour()

	// 深夜帯のアクセス
	if hour >= 0 && hour <= 6 {
		base.Events[0].Name = "suspicious_access"
		base.IPAddress = generateSuspiciousIP(rng)
	}

	return base
}

// 大量アクセス異常
func generateVolumeAnomaly(base *GoogleWorkspaceLogEntry, rng *rand.Rand) *GoogleWorkspaceLogEntry {
	// 大量ダウンロードパターン
	if len(base.Events) > 0 {
		base.Events[0].Name = "bulk_download"
		for i := range base.Events[0].Parameters {
			if base.Events[0].Parameters[i].Name == "doc_type" {
				base.Events[0].Parameters[i].Value = "file"
			}
		}
	}

	return base
}

// 疑わしいIPアドレスを生成
func generateSuspiciousIP(rng *rand.Rand) string {
	suspiciousIPs := []string{
		"198.51.100.10", // US IP
		"203.0.113.45",  // External partner
		"192.0.2.100",   // Test IP
	}
	return suspiciousIPs[rng.Intn(len(suspiciousIPs))]
}

// 外部IPアドレスを生成
func generateExternalIP(rng *rand.Rand) string {
	externalIPs := []string{
		"203.0.113.45",  // 外部からのアクセス
		"198.51.100.10", // パートナー企業
		"192.0.2.100",   // テストIP
	}
	return externalIPs[rng.Intn(len(externalIPs))]
}

// 内部IPアドレスを生成
func generateInternalIP(rng *rand.Rand) string {
	return fmt.Sprintf("192.168.1.%d", rng.Intn(254)+1)
}

// VPN経由の内部IPアドレスを生成
func generateVPNInternalIP(rng *rand.Rand) string {
	return fmt.Sprintf("10.0.100.%d", rng.Intn(254)+1)
}

// 外部プロファイルIDを生成
func generateExternalProfileID(rng *rand.Rand) string {
	return fmt.Sprintf("external_%d", rng.Int63())
}

// ドキュメントIDを生成
func generateDocumentID(rng *rand.Rand) string {
	return fmt.Sprintf("1%s", generateRandomString(rng, 43))
}

// ランダムデータセット名を生成
func generateRandomDataset(rng *rand.Rand) string {
	datasets := []string{
		"202408",
		"Q2_2024",
		"final_exam",
		"midterm_results",
		"progress_report",
	}
	return datasets[rng.Intn(len(datasets))]
}

// ランダム文字列を生成
func generateRandomString(rng *rand.Rand, length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rng.Intn(len(chars))]
	}
	return string(result)
}
