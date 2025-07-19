# Log Generator (loggen) 仕様書 v2.0

## 概要

セキュリティキャンプ2025 B1講義用のGoogle Workspace監査ログ生成システム。**ログシード**（時刻+決定論的パラメータ）を事前生成し、リクエスト時に演繹的にログ内容を生成する軽量・高効率なアプローチを採用。

## 設計思想

- **軽量シード**: 全ログの生データではなく、生成に必要な最小限の情報のみ保存
- **演繹生成**: シードから決定論的にログ内容を再構築
- **高頻度対応**: 毎秒10件のログ生成に対応（864,000件/日）
- **容量効率**: 1日分864,000件のログを数MBのシードで実現
- **冪等性保証**: 同一シードから必ず同一のログを生成

## アーキテクチャ

### データフロー
```
tools/loggen → ログシード定義 → 共有ライブラリ → Lambda演繹生成 → API配信
```

### 責務分離
- **loggen**: ログシード生成・検証・分析ツール
- **共有ライブラリ**: シード→ログ変換ロジック（loggen, Lambda両方で使用）
- **Lambda**: シード読み込み・演繹生成・API配信

### ディレクトリ構造
```
internal/logcore/           # 共有ライブラリ（リポジトリ内パッケージ）
├── types.go               # 共通型定義（LogSeed, DayTemplate等）
├── generator.go           # シード→ログ変換ロジック
├── config.go              # 設定・マスターデータ
└── anomaly.go             # 異常パターン実装

tools/loggen/
├── spec_v2.md             # 本仕様書
├── main.go                # メインエントリーポイント
├── cmd/
│   ├── generate.go        # シード生成コマンド
│   ├── validate.go        # シード検証コマンド
│   └── preview.go         # ログプレビューコマンド
├── internal/
│   ├── seed/
│   │   ├── generator.go   # シード生成ロジック
│   │   ├── timeline.go    # 時系列設計
│   │   └── anomaly.go     # 異常パターン配置
│   └── output/
│       ├── manifest.go    # シードマニフェスト出力
│       └── compress.go    # 圧縮処理
├── testdata/
│   ├── expected/          # 期待値データ
│   └── samples/           # サンプル出力
└── output/                # 生成結果（gitignore）
    └── seeds/
        ├── day_template.json.gz    # 1日分のシードテンプレート
        └── anomaly_patterns.json  # 異常パターン定義

terraform/lambda/auditlog/
├── main.go                # Lambda エントリーポイント
├── internal/
│   ├── api/
│   │   ├── handler.go     # HTTP ハンドラー
│   │   └── response.go    # レスポンス生成
│   └── embed/
│       └── seeds.go       # シードファイルのembed
└── seeds/                 # tools/loggen/output/seeds への参照
    ├── day_template.json.gz
    └── anomaly_patterns.json
```

## 機能仕様

### 1. ログシード設計

#### 1.1 シード構造
```go
// 1日分のログシード（864,000件）
type DayTemplate struct {
    Date        string      `json:"date"`        // "2024-08-12"
    LogSeeds    []LogSeed   `json:"log_seeds"`   // 864,000個のシード
    Metadata    SeedMeta    `json:"metadata"`
}

type LogSeed struct {
    Timestamp   int64    `json:"ts"`       // Unix秒 (相対時刻)
    EventType   uint8    `json:"et"`       // イベント種別ID
    UserIndex   uint8    `json:"ui"`       // ユーザーインデックス
    ResourceIdx uint8    `json:"ri"`       // リソースインデックス  
    Pattern     uint8    `json:"pt"`       // 正常(0)/異常(1-10)パターン
    Seed        uint32   `json:"seed"`     // この秒のランダムシード
}

type SeedMeta struct {
    TotalLogs    int                 `json:"total_logs"`
    NormalRatio  float64            `json:"normal_ratio"`
    AnomalyStats map[string]int     `json:"anomaly_stats"`
    Generated    time.Time          `json:"generated"`
}
```

#### 1.2 容量効率化
```go
// 1ログあたり約10バイト × 864,000件 = 約8.64MB
// gzip圧縮で約2-3MB/日の見込み

// さらなる効率化: 時間帯別の差分圧縮
type HourlyTemplate struct {
    Hour        uint8       `json:"hour"`      // 0-23
    BasePattern uint8       `json:"base"`      // この時間の基本パターン
    Variations  []LogSeed   `json:"vars"`      // 差分のみ記録
}
```

### 2. 演繹的ログ生成

#### 2.1 共有ライブラリ設計
```go
// internal/logcore/generator.go
package logcore

// Google Workspace監査ログエントリ構造
type GoogleWorkspaceLogEntry struct {
    Kind       string     `json:"kind"`
    ID         LogID      `json:"id"`
    Actor      Actor      `json:"actor"`
    OwnerDomain string    `json:"ownerDomain"`
    IPAddress  string     `json:"ipAddress"`
    Events     []Event    `json:"events"`
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
    Name         string   `json:"name"`
    Value        string   `json:"value,omitempty"`
    BoolValue    bool     `json:"boolValue,omitempty"`
    MultiStrValue []string `json:"multiStrValue,omitempty"`
}

// シードからログエントリを生成
func GenerateLogEntry(seed LogSeed, baseDate time.Time, config *Config) *GoogleWorkspaceLogEntry {
    // 1. タイムスタンプ計算
    timestamp := baseDate.Add(time.Duration(seed.Timestamp) * time.Second)
    
    // 2. シードベースの疑似乱数生成器
    rng := rand.New(rand.NewSource(int64(seed.Seed)))
    
    // 3. ユーザー・リソース解決
    user := config.Users[seed.UserIndex]
    resource := config.Resources[seed.ResourceIdx]
    
    // 4. 基本ログエントリ作成
    logEntry := &GoogleWorkspaceLogEntry{
        Kind: "audit#activity",
        ID: LogID{
            Time:            timestamp.Format(time.RFC3339Nano),
            UniqueQualifier: generateUniqueQualifier(rng),
            ApplicationName: determineApplicationName(seed.EventType),
            CustomerID:      config.CustomerID,
        },
        Actor: Actor{
            CallerType: "USER",
            Email:      user.Email,
            ProfileID:  user.ProfileID,
        },
        OwnerDomain: config.OwnerDomain,
        IPAddress:   generateIPAddress(user, rng),
        Events:      []Event{},
    }
    
    // 5. イベント種別による生成分岐
    switch seed.EventType {
    case EventTypeDriveAccess:
        logEntry.Events = append(logEntry.Events, generateDriveAccessEvent(user, resource, rng))
    case EventTypeLogin:
        logEntry.Events = append(logEntry.Events, generateLoginEvent(user, timestamp, rng))
    case EventTypeAdmin:
        logEntry.Events = append(logEntry.Events, generateAdminEvent(user, resource, rng))
    }
    
    // 6. 異常パターンの適用
    if seed.Pattern > 0 {
        logEntry = applyAnomalyPattern(logEntry, seed.Pattern, rng)
    }
    
    return logEntry
}
```

#### 2.2 異常パターン実装
```go
// internal/logcore/anomaly.go

type AnomalyGenerator struct {
    Patterns map[uint8]AnomalyFunc
}

type AnomalyFunc func(*LogEntry, *rand.Rand) *LogEntry

func NewAnomalyGenerator() *AnomalyGenerator {
    return &AnomalyGenerator{
        Patterns: map[uint8]AnomalyFunc{
            1: generateExample1NightdownAdminDownload, // 実例1: 夜間管理者による大量データダウンロード
            2: generateExample2ExternalLinkAccess,      // 実例2: anyone with link設定ミスによる外部流出
            3: generateExample3VpnLateralMovement,      // 実例3: VPN脆弱性経由の水平移動攻撃
            4: generateTimeAnomaly,                     // 時間外アクセス
            5: generateVolumeAnomaly,                   // 大量アクセス
        },
    }
}

func generateExample1NightdownAdminDownload(base *GoogleWorkspaceLogEntry, rng *rand.Rand) *GoogleWorkspaceLogEntry {
    // 実例1: 夜間の管理者による大量学習データダウンロード
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

func generateExample2ExternalLinkAccess(base *GoogleWorkspaceLogEntry, rng *rand.Rand) *GoogleWorkspaceLogEntry {
    // 実例2: anyone with link設定ミスによる機密情報の意図しない外部流出
    
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

func generateExample3VpnLateralMovement(base *GoogleWorkspaceLogEntry, rng *rand.Rand) *GoogleWorkspaceLogEntry {
    // 実例3: VPN脆弱性経由の攻撃 - Google Workspaceでの異常アクセス試行
    
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
```

### 3. tools/loggen の実装

#### 3.1 シード生成戦略
```go
// tools/loggen/internal/seed/generator.go

func GenerateDayTemplate(date time.Time) (*DayTemplate, error) {
    template := &DayTemplate{
        Date:     date.Format("2006-01-02"),
        LogSeeds: make([]LogSeed, 0, 864000),
    }
    
    // 1秒ごとにログシードを生成
    for second := 0; second < 86400; second++ {
        currentTime := date.Add(time.Duration(second) * time.Second)
        
        // この秒のログ件数決定（平均10件、ポワソン分布）
        logCount := generateLogCount(currentTime)
        
        for i := 0; i < logCount; i++ {
            seed := LogSeed{
                Timestamp: int64(second),
                Seed:      generateSecondSeed(currentTime, i),
            }
            
            // 時間帯・曜日による活動パターン決定
            seed.EventType, seed.UserIndex, seed.ResourceIdx = 
                selectActivityPattern(currentTime)
            
            // 異常パターンの配置判定
            seed.Pattern = determineAnomalyPattern(currentTime, i)
            
            template.LogSeeds = append(template.LogSeeds, seed)
        }
    }
    
    return template, nil
}

func generateLogCount(t time.Time) int {
    hour := t.Hour()
    weekday := t.Weekday()
    
    // 時間帯・曜日による期待値調整
    expectedRate := getExpectedRate(hour, weekday)
    
    // ポワソン分布でログ件数決定
    return poissonRandom(expectedRate)
}

func getExpectedRate(hour int, weekday time.Weekday) float64 {
    // 基本レート: 毎秒10件
    baseRate := 10.0
    
    // 時間帯補正
    timeMultiplier := map[int]float64{
        0: 0.1, 1: 0.05, 2: 0.05, 3: 0.05, 4: 0.05, 5: 0.1,  // 深夜
        6: 0.2, 7: 0.4, 8: 0.8,                               // 朝
        9: 1.2, 10: 1.5, 11: 1.8, 12: 1.0,                   // 午前〜昼
        13: 0.8, 14: 1.3, 15: 1.6, 16: 1.4, 17: 1.1, 18: 0.9, // 午後
        19: 0.6, 20: 0.4, 21: 0.3, 22: 0.2, 23: 0.15,        // 夜
    }
    
    // 曜日補正
    weekdayMultiplier := map[time.Weekday]float64{
        time.Monday: 1.0, time.Tuesday: 1.1, time.Wednesday: 1.2,
        time.Thursday: 1.1, time.Friday: 0.9, 
        time.Saturday: 0.3, time.Sunday: 0.2,
    }
    
    return baseRate * timeMultiplier[hour] * weekdayMultiplier[weekday]
}
```

#### 3.2 異常パターン配置戦略
```go
// tools/loggen/internal/seed/anomaly.go

func determineAnomalyPattern(t time.Time, sequenceInSecond int) uint8 {
    hour := t.Hour()
    
    // 実例1: 夜間の管理者による大量データダウンロード
    if (hour >= 18 || hour <= 9) && rand.Float64() < 0.12 {
        return 1 // Example1 pattern
    }
    
    // 実例2: anyone with link設定ミスによる外部流出（まとまって発生）
    if hour >= 10 && hour <= 16 {
        // 15分間隔でバーストパターン（外部からの自動アクセス）
        if t.Minute()%15 < 3 && rand.Float64() < 0.25 {
            return 2 // Example2 pattern
        }
    }
    
    // 実例3: VPN脆弱性経由の水平移動攻撃（業務時間内の怪しい動作）
    if hour >= 9 && hour <= 18 && rand.Float64() < 0.08 {
        return 3 // Example3 pattern
    }
    
    // 一般的な軽微異常
    if rand.Float64() < 0.08 {
        return uint8(4 + rand.Intn(2)) // Pattern 4-5
    }
    
    return 0 // 正常パターン
}
```

### 4. Lambda実装

#### 4.1 API Handler
```go
// terraform/lambda/auditlog/internal/api/handler.go

type LogHandler struct {
    DayTemplate  *DayTemplate
    Config       *logcore.Config
    Generator    *logcore.Generator
}

func (h *LogHandler) HandleRequest(startTime, endTime time.Time) ([]byte, error) {
    var logs []logcore.LogEntry
    
    // 要求された時間範囲のシードを抽出
    seeds := h.extractSeedsInRange(startTime, endTime)
    
    // 各シードからログエントリを生成
    for _, seed := range seeds {
        logEntry := h.Generator.GenerateLogEntry(seed, startTime.Truncate(24*time.Hour), h.Config)
        logs = append(logs, *logEntry)
    }
    
    // JSONL形式で出力
    return h.convertToJSONL(logs), nil
}

func (h *LogHandler) extractSeedsInRange(start, end time.Time) []logcore.LogSeed {
    startSecond := int64(start.Hour()*3600 + start.Minute()*60 + start.Second())
    endSecond := int64(end.Hour()*3600 + end.Minute()*60 + end.Second())
    
    var seeds []logcore.LogSeed
    for _, seed := range h.DayTemplate.LogSeeds {
        if seed.Timestamp >= startSecond && seed.Timestamp < endSecond {
            seeds = append(seeds, seed)
        }
    }
    
    return seeds
}
```

#### 4.2 日付調整機能
```go
// terraform/lambda/auditlog/internal/api/handler.go

func (h *LogHandler) AdjustDateToToday(requestTime time.Time) time.Time {
    // テンプレートは固定日付 (2024-08-12) で作成
    templateDate := time.Date(2024, 8, 12, 0, 0, 0, 0, time.UTC)
    
    // リクエスト日付にテンプレートの時刻パターンを適用
    adjustedTime := time.Date(
        requestTime.Year(), requestTime.Month(), requestTime.Day(),
        requestTime.Hour(), requestTime.Minute(), requestTime.Second(), requestTime.Nanosecond(),
        requestTime.Location(),
    )
    
    return adjustedTime
}
```

### 5. CLI仕様

#### 5.1 シード生成
```bash
# 1日分のシードテンプレート生成
./loggen generate --date 2024-08-12 --output ./output/seeds/

# 異常パターン比率調整
./loggen generate --anomaly-ratio 0.20 --output ./output/seeds/

# プレビュー機能（実際のログ生成）
./loggen preview --seeds ./output/seeds/day_template.json.gz --time-range "10:00-11:00"
```

#### 5.2 検証・分析
```bash
# シード検証
./loggen validate --seeds ./output/seeds/day_template.json.gz

# 統計情報
./loggen stats --seeds ./output/seeds/day_template.json.gz

# 異常パターン分析
./loggen analyze --seeds ./output/seeds/day_template.json.gz --pattern-type phase3
```

### 6. 容量・パフォーマンス

#### 6.1 容量効率
```
1日分のログ生成に必要なデータ:
- LogSeed: 10バイト × 864,000件 = 8.64MB
- gzip圧縮後: 約2-3MB
- メタデータ: 数KB
- 設定データ: 10-20KB

総容量: 約3-4MB/日
30日分でも: 約100-120MB
```

#### 6.2 パフォーマンス
```
Lambda冷起動時:
- シード読み込み: 50-100ms
- 設定読み込み: 10-20ms

リクエスト処理時:
- 1時間分(36,000件): 200-500ms
- シード抽出: 10-50ms
- ログ生成: 150-400ms
- JSONL変換: 20-50ms

目標レスポンス時間: 1秒以内
```

### 7. 共有ライブラリの参照

#### 7.1 Go Modules による内部パッケージ参照
```go
// tools/loggen/cmd/generate.go
import (
    "github.com/m-mizutani/seccamp-2025-b1/internal/logcore"
)

// terraform/lambda/auditlog/internal/api/handler.go  
import (
    "github.com/m-mizutani/seccamp-2025-b1/internal/logcore"
)
```

#### 7.2 シードファイル参照
```bash
# tools/loggen でシード生成
./loggen generate --output ./output/seeds/

# terraform/lambda/auditlog で参照（embed）
# シードファイルは相対パスで参照
```

#### 7.2 バージョン管理
```go
// 共有ライブラリのバージョン管理
const (
    LogCoreVersion = "v1.0.0"
    SeedFormatVersion = "v1.0.0"
)

// バージョン不整合の検出
func ValidateCompatibility(seedMeta SeedMeta) error {
    if seedMeta.LogCoreVersion != LogCoreVersion {
        return fmt.Errorf("logcore version mismatch: seed=%s, lib=%s", 
            seedMeta.LogCoreVersion, LogCoreVersion)
    }
    return nil
}
```

---

## 実装優先度

### Phase 1（基本機能）
1. LogSeed構造定義・基本的なシード生成
2. 共有ライブラリの基本的な演繹生成機能
3. Lambda基本API・日付調整機能
4. 正常系ログパターンの実装

### Phase 2（異常系強化）
1. 無敗塾実例1〜3の異常パターン実装
2. 時間帯・曜日別の現実的な分散
3. 異常パターンの配置戦略精密化
4. 包括的な検証・分析機能

### Phase 3（最適化・運用）
1. 容量・パフォーマンス最適化
2. 豊富なCLI機能・統計表示
3. エラーハンドリング・ロバスト性向上
4. 包括的なテスト・ドキュメント


## 実装制約

- エラー処理には github.com/m-mizutani/goerr/v2 を使う
- CLI制御には github.com/urfave/cli/v3 を使う
- テストは github.com/m-mizutani/gt を使う

---


**作成日**: 2024年7月14日  
**対象**: セキュリティキャンプ2025 B1 ログ生成システム v2.0  
**更新者**: Claude Code Assistant