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
			PatternExample1NightAdminDownload:    generateExample1NightAdminDownload,
			PatternExample2ExternalLinkAccess:    generateExample2ExternalLinkAccess,
			PatternExample3VpnLateralMovement:    generateExample3VpnLateralMovement,
			PatternTimeAnomaly:                   generateTimeAnomaly,
			PatternVolumeAnomaly:                 generateVolumeAnomaly,
			PatternExample4HighFreqAuthAttack:    generateExample4HighFreqAuthAttack,
			PatternExample5RapidDataTheft:        generateExample5RapidDataTheft,
			PatternExample6MultiServiceProbing:   generateExample6MultiServiceProbing,
			PatternExample7SimultaneousGeoAccess: generateExample7SimultaneousGeoAccess,
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

// Pattern 4: 高頻度認証攻撃
func generateExample4HighFreqAuthAttack(base *GoogleWorkspaceLogEntry, rng *rand.Rand) *GoogleWorkspaceLogEntry {
	// 固定の攻撃者IP
	base.IPAddress = "203.0.113.99"
	
	// ログインイベントに変更
	base.ID.ApplicationName = "login"
	
	// 95%は失敗、5%は成功
	if rng.Float32() < 0.95 {
		base.Events = []Event{
			{
				Type: "login",
				Name: "login_failure",
				Parameters: []Parameter{
					{Name: "login_type", Value: "google_password"},
					{Name: "login_failure_type", Value: "account_disabled"},
					{Name: "is_suspicious", BoolValue: true},
				},
			},
		}
	} else {
		// 成功した場合
		base.Events = []Event{
			{
				Type: "login",
				Name: "login_success",
				Parameters: []Parameter{
					{Name: "login_type", Value: "google_password"},
					{Name: "login_challenge_method", MultiStrValue: []string{"password"}},
					{Name: "is_suspicious", BoolValue: true},
				},
			},
		}
	}
	
	return base
}

// Pattern 5: 超高速データ窃取
func generateExample5RapidDataTheft(base *GoogleWorkspaceLogEntry, rng *rand.Rand) *GoogleWorkspaceLogEntry {
	// 特定の侵害されたユーザーとIP
	base.Actor.Email = "tanaka.hiroshi@muhaijuku.com"
	base.IPAddress = "198.51.100.99"
	
	// ドライブアクセスイベント
	base.ID.ApplicationName = "drive"
	
	// ランダムなファイルへの大量ダウンロード
	confidentialFiles := []string{
		"財務報告書_2024Q3.xlsx",
		"顧客リスト_機密.csv",
		"製品開発計画_2025.docx",
		"人事評価データ.xlsx",
		"研究データ_機密.zip",
	}
	
	base.Events = []Event{
		{
			Type: "access",
			Name: "download",
			Parameters: []Parameter{
				{Name: "doc_id", Value: generateDocumentID(rng)},
				{Name: "doc_title", Value: confidentialFiles[rng.Intn(len(confidentialFiles))]},
				{Name: "doc_type", Value: "file"},
				{Name: "owner", Value: "admin@muhai-academy.com"},
				{Name: "visibility", Value: "private"},
				{Name: "primary_event", BoolValue: true},
				{Name: "size_bytes", Value: fmt.Sprintf("%d", 1000000+rng.Intn(50000000))}, // 1MB-50MB
			},
		},
	}
	
	return base
}

// Pattern 6: マルチサービス不正アクセス試行
func generateExample6MultiServiceProbing(base *GoogleWorkspaceLogEntry, rng *rand.Rand) *GoogleWorkspaceLogEntry {
	// 特定の感染ユーザー
	base.Actor.Email = "sato.yuki@muhaijuku.com"
	
	// サービスをランダムに選択
	services := []string{"drive", "calendar", "gmail", "admin"}
	selectedService := services[rng.Intn(len(services))]
	base.ID.ApplicationName = selectedService
	
	// 80%は権限エラー（検知閾値70%を確実に超えるため）
	if rng.Float32() < 0.8 {
		switch selectedService {
		case "drive":
			base.Events = []Event{
				{
					Type: "access",
					Name: "access_denied",
					Parameters: []Parameter{
						{Name: "doc_id", Value: generateDocumentID(rng)},
						{Name: "doc_title", Value: "機密ファイル"},
						{Name: "denied_reason", Value: "insufficient_permissions"},
					},
				},
			}
		case "admin":
			base.Events = []Event{
				{
					Type: "USER_SETTINGS",
					Name: "PERMISSION_DENIED",
					Parameters: []Parameter{
						{Name: "USER_EMAIL", Value: base.Actor.Email},
						{Name: "denied_reason", Value: "not_admin"},
					},
				},
			}
		default:
			base.Events = []Event{
				{
					Type: "access",
					Name: "permission_denied",
					Parameters: []Parameter{
						{Name: "service", Value: selectedService},
						{Name: "denied_reason", Value: "unauthorized_access"},
					},
				},
			}
		}
	} else {
		// 20%は成功（プロービングの一環として一部成功）
		switch selectedService {
		case "drive":
			base.Events = []Event{
				{
					Type: "access",
					Name: "view",
					Parameters: []Parameter{
						{Name: "doc_id", Value: generateDocumentID(rng)},
						{Name: "doc_title", Value: "共有ドキュメント"},
					},
				},
			}
		case "calendar":
			base.Events = []Event{
				{
					Type: "event_change",
					Name: "view_event",
					Parameters: []Parameter{
						{Name: "calendar_id", Value: "shared"},
						{Name: "event_id", Value: fmt.Sprintf("%s", generateRandomString(rng, 12))},
					},
				},
			}
		case "gmail":
			base.Events = []Event{
				{
					Type: "mail_action",
					Name: "list_messages",
					Parameters: []Parameter{
						{Name: "folder", Value: "INBOX"},
					},
				},
			}
		default:
			base.Events = []Event{
				{
					Type: "access",
					Name: "read",
					Parameters: []Parameter{
						{Name: "service", Value: selectedService},
					},
				},
			}
		}
	}
	
	return base
}

// Pattern 7: 地理的同時アクセス
func generateExample7SimultaneousGeoAccess(base *GoogleWorkspaceLogEntry, rng *rand.Rand) *GoogleWorkspaceLogEntry {
	// 特定のユーザー
	base.Actor.Email = "yamada.takeshi@muhaijuku.com"
	
	// 50%の確率で国を切り替え
	if rng.Float32() < 0.5 {
		// 日本からのアクセス
		base.IPAddress = "192.0.2.10"
	} else {
		// 米国からのアクセス
		base.IPAddress = "198.51.100.20"
	}
	
	// 通常の業務操作
	operations := []struct {
		appName string
		event   Event
	}{
		{
			appName: "drive",
			event: Event{
				Type: "access",
				Name: "view",
				Parameters: []Parameter{
					{Name: "doc_id", Value: generateDocumentID(rng)},
					{Name: "doc_title", Value: "業務ファイル.docx"},
					{Name: "doc_type", Value: "document"},
				},
			},
		},
		{
			appName: "gmail",
			event: Event{
				Type: "mail_action",
				Name: "send_message",
				Parameters: []Parameter{
					{Name: "message_id", Value: fmt.Sprintf("<%s@mail.gmail.com>", generateRandomString(rng, 16))},
					{Name: "recipient", Value: "colleague@muhai-academy.com"},
				},
			},
		},
	}
	
	selected := operations[rng.Intn(len(operations))]
	base.ID.ApplicationName = selected.appName
	base.Events = []Event{selected.event}
	
	return base
}
