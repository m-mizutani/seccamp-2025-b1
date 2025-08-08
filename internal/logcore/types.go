package logcore

import "time"

const (
	LogCoreVersion    = "v1.0.0"
	SeedFormatVersion = "v1.0.0"
)

// 1日分のログシード（864,000件）
type DayTemplate struct {
	Date     string    `json:"date"`      // "2024-08-12"
	LogSeeds []LogSeed `json:"log_seeds"` // 864,000個のシード
	Metadata SeedMeta  `json:"metadata"`
}

// ログシード構造（約10バイト）
type LogSeed struct {
	Timestamp   int64  `json:"ts"`   // Unix秒 (相対時刻)
	EventType   uint8  `json:"et"`   // イベント種別ID
	UserIndex   uint8  `json:"ui"`   // ユーザーインデックス
	ResourceIdx uint8  `json:"ri"`   // リソースインデックス
	Pattern     uint8  `json:"pt"`   // 正常(0)/異常(1-10)パターン
	Seed        uint32 `json:"seed"` // この秒のランダムシード
}

// シードメタデータ
type SeedMeta struct {
	TotalLogs         int            `json:"total_logs"`
	NormalRatio       float64        `json:"normal_ratio"`
	AnomalyStats      map[string]int `json:"anomaly_stats"`
	Generated         time.Time      `json:"generated"`
	LogCoreVersion    string         `json:"logcore_version"`
	SeedFormatVersion string         `json:"seed_format_version"`
}

// Google Workspace監査ログエントリ構造
type GoogleWorkspaceLogEntry struct {
	Kind        string  `json:"kind"`
	ID          LogID   `json:"id"`
	Actor       Actor   `json:"actor"`
	OwnerDomain string  `json:"ownerDomain"`
	IPAddress   string  `json:"ipAddress"`
	Events      []Event `json:"events"`
}

type LogID struct {
	Time            string `json:"time"`
	UniqueQualifier string `json:"uniqueQualifier"`
	ApplicationName string `json:"applicationName"`
	CustomerID      string `json:"customerId"`
}

type Actor struct {
	CallerType string `json:"callerType"`
	Email      string `json:"email"`
	ProfileID  string `json:"profileId"`
}

type Event struct {
	Type       string      `json:"type"`
	Name       string      `json:"name"`
	Parameters []Parameter `json:"parameters"`
}

type Parameter struct {
	Name          string   `json:"name"`
	Value         string   `json:"value,omitempty"`
	BoolValue     bool     `json:"boolValue,omitempty"`
	MultiStrValue []string `json:"multiStrValue,omitempty"`
}

// 設定用の構造体
type User struct {
	Email     string `json:"email"`
	ProfileID string `json:"profile_id"`
	Role      string `json:"role"`
}

type Resource struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	ID         string `json:"id"`
	Visibility string `json:"visibility"`
}

type Config struct {
	CustomerID  string     `json:"customer_id"`
	OwnerDomain string     `json:"owner_domain"`
	Users       []User     `json:"users"`
	Resources   []Resource `json:"resources"`
	BaseDate    time.Time  `json:"base_date"`
}

// イベント種別定数
const (
	EventTypeDriveAccess uint8 = 1
	EventTypeLogin       uint8 = 2
	EventTypeAdmin       uint8 = 3
	EventTypeCalendar    uint8 = 4
	EventTypeGmail       uint8 = 5
)

// 異常パターン定数
const (
	PatternNormal                     uint8 = 0
	PatternExample1NightAdminDownload uint8 = 1
	PatternExample2ExternalLinkAccess uint8 = 2
	PatternExample3VpnLateralMovement uint8 = 3
	PatternTimeAnomaly                uint8 = 4
	PatternVolumeAnomaly              uint8 = 5
)
