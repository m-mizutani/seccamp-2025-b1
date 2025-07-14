# Google Workspace 疑似監査ログ生成 Lambda 仕様書

## 概要

セキュリティキャンプ2025 B1講義「Cloud Platform Monitoring Introduction」の実習で使用するGoogle Workspace監査ログの疑似データを生成・配信するLambda関数。無敗塾（Muhaijuku）のシナリオに基づいた現実的な監査ログデータを提供する。

## 機能要件

### 1. HTTPエンドポイント

#### エンドポイント
- **URL**: `https://api.workshop.example.com/logs` (実際はLambdaが発行するURLがbase URLになる)
- **Method**: GET
- **認証**: なし

#### リクエストパラメータ
```
GET /logs?start_time=2024-08-12T10:00:00Z&end_time=2024-08-12T11:00:00Z
```

| パラメータ | 必須 | 形式 | 説明 |
|-----------|------|------|------|
| start_time | Yes | ISO8601 | ログ取得開始時刻 |
| end_time | Yes | ISO8601 | ログ取得終了時刻 |
| offset | No | Number | default=0 |
| limit | No | Number | default=50 |

#### レスポンス
- **Content-Type**: `application/json`
- **Format**: JSONL（JSON Lines）- 1行に1つのログエントリ
- **Compression**: gzip（オプション、Accept-Encoding: gzipの場合）

#### 制約
- **データ保持期間**: 過去30日分のログデータ

### 2. ログデータ形式

#### Google Workspace監査ログ形式
```json
{
  "id": "unique-log-id",
  "timestamp": "2024-08-12T10:15:30Z",
  "user": {
    "email": "teacher@muhaijuku.com",
    "name": "田中太郎",
    "domain": "muhaijuku.com"
  },
  "event": {
    "type": "drive",
    "name": "access",
    "action": "view"
  },
  "resource": {
    "name": "grades/math_test_results.xlsx",
    "id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
    "type": "file"
  },
  "metadata": {
    "ip_address": "192.168.1.100",
    "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    "location": {
      "country": "Japan",
      "region": "Tokyo",
      "city": "Shibuya"
    }
  },
  "result": {
    "success": true,
    "denied_reason": null
  }
}
```

### 3. 無敗塾シナリオベースのログパターン

#### 3.1 正常パターン

##### 通常業務時間内のファイルアクセス
- **時間帯**: 平日 09:00-18:00
- **ユーザー**: `@muhaijuku.com` ドメイン
- **リソース**:
  - 教材フォルダ (`materials/`)
  - 一般管理資料 (`admin/general/`)
  - 個人フォルダ (`personal/{user}/`)

##### 正常な権限変更操作
- **操作者**: 管理者ユーザー
- **対象**: フォルダ共有設定
- **時間帯**: 業務時間内

##### 通常範囲内のログイン活動
- **頻度**: 1日1-3回程度
- **場所**: 日本国内
- **時間帯**: 業務時間±2時間

#### 3.2 異常パターン（無敗塾フェーズ別）

##### フェーズ3: 試験期間外成績データアクセス
```json
{
  "event": { "name": "access_denied" },
  "resource": { "name": "grades/math_test_results.xlsx" },
  "user": { "email": "teacher@muhaijuku.com" },
  "timestamp": "2024-08-12T22:30:00Z",
  "result": { "success": false, "denied_reason": "insufficient_permissions" }
}
```
- **パターン**: 1時間で5回以上の`access_denied`
- **対象**: `grades/` フォルダ内ファイル
- **時間**: 業務時間外（19:00-08:00）

##### フェーズ4: 外部ドメインからの大量ファイルアクセス
```json
{
  "user": { "email": "external.instructor@partner-company.com", "domain": "partner-company.com" },
  "event": { "name": "access", "action": "download" },
  "resource": { "name": "ai_training_data/student_essays.zip" },
  "metadata": { "ip_address": "203.0.113.45" }
}
```
- **パターン**: 外部ドメインユーザーによる機密データアクセス
- **頻度**: 短時間で大量ファイルアクセス（10件/10分）

##### フェーズ5: 海外からの管理者権限使用
```json
{
  "user": { "email": "admin@muhaijuku.com" },
  "event": { "name": "admin_settings_change" },
  "metadata": {
    "ip_address": "198.51.100.10",
    "location": { "country": "United States", "city": "New York" }
  },
  "timestamp": "2024-08-12T02:00:00Z"
}
```
- **パターン**: 地理的に異常な場所からの特権操作
- **時間**: 日本時間の深夜帯

### 4. データ生成ロジック

#### 4.1 時系列データの整合性
- **基準時間**: JST（UTC+9）
- **ログ頻度**: 正常時 5-15件/時間、異常時 20-50件/時間
- **時間分散**: ログ発生時刻を現実的に分散（完全にランダムではない）

#### 4.2 ユーザー・リソースの一貫性
- **ユーザー**: 10-15名の固定ユーザープール
- **ファイル**: 階層化されたフォルダ構造（無敗塾の組織に対応）
- **IPアドレス**: 範囲を限定した現実的なIP（プライベート・パブリック）

#### 4.3 異常データの混在比率
- **正常ログ**: 80-90%
- **軽微な異常**: 8-15%（単発のアクセス拒否等）
- **重要な異常**: 2-5%（連続アクセス拒否、外部アクセス等）

### 5. 技術実装仕様

#### 5.1 Lambda関数構成
- **Runtime**: Go（provided.al2）
- **Architecture**: arm64
- **Memory**: 256MB（埋め込みデータのため軽量）
- **Timeout**: 10秒
- **Concurrent Executions**: 10

#### 5.2 データストレージ方式
**事前生成ログのembedによる配信**

- **ログデータ**: Goバイナリにembedされた静的JSONL
- **時間範囲**: 30日分の時間別ログファイル（720ファイル）
- **ファイル構成**: `logs/2024/08/12/10.jsonl.gz` 形式
- **冪等性**: ファイルベースのため完全に保証

#### 5.3 環境変数
```bash
TIMEZONE=Asia/Tokyo
LOG_RETENTION_DAYS=30
RATE_LIMIT_PER_MINUTE=10
DEFAULT_DOMAIN=muhaijuku.com
EXTERNAL_DOMAINS=partner-company.com,consulting-firm.jp
```

#### 5.4 エラーレスポンス
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Try again in 60 seconds.",
    "retry_after": 60
  }
}
```

**エラーコード**:
- `INVALID_TIME_RANGE`: 時間範囲が無効
- `TIME_RANGE_TOO_LARGE`: 1時間を超える範囲指定
- `RATE_LIMIT_EXCEEDED`: レート制限超過
- `INVALID_TOKEN`: 認証トークン無効
- `INTERNAL_ERROR`: 内部エラー

### 6. セキュリティ要件

#### 6.1 認証・認可
- **認証方式**: OAuth 2.0 Bearer Token
- **トークン検証**: ハードコードされた有効トークンリスト
- **スコープ**: 読み取り専用

#### 6.2 ログ出力制限
- **個人情報**: 実際の個人情報は含まない（仮名のみ）
- **機密情報**: 実際の機密データは含まない
- **IPアドレス**: 実際の組織IPは含まない

### 7. 運用要件

#### 7.1 監視・ロギング
- **CloudWatch Logs**: Lambda実行ログ
- **CloudWatch Metrics**: レート制限メトリクス、エラー率
- **X-Ray**: リクエストトレーシング（オプション）

#### 7.2 設定変更
- **異常パターン**: S3設定ファイルで動的変更可能
- **ユーザー/リソース**: 環境変数で調整可能
- **レート制限**: 環境変数で調整可能

### 8. テスト要件

#### 8.1 単体テスト
- 時間範囲指定の検証
- レート制限の動作確認
- 異常パターン生成の確認

#### 8.2 統合テスト
- 受講生Lambda（課題1）からの接続テスト
- 長期間のログ取得シミュレーション
- エラーハンドリングの確認

### 9. パフォーマンス要件

#### 9.1 レスポンス時間
- **通常時**: 1秒以内
- **最大時**: 3秒以内（1時間分フルデータ）

#### 9.2 スループット
- **同時接続**: 最大10接続
- **データ量**: 1時間分で最大1MB

---

## 実装制約・冪等性保証

### 冪等性要件
**同一の取得条件（start_time, end_time, offset, limit）では、必ず同じログセットを返すこと**

APIの責務として、受講生が何度同じ条件でログを取得しても、完全に同一のログデータが取得できる必要がある。これにより：
- 受講生のLambda実装でリトライ処理を安全に実装可能
- 講師による動作確認が再現可能
- デバッグ・トラブルシューティングが容易

### 冪等性実装方法

#### 1. 決定論的ログ生成
**疑似乱数生成器のシード値固定**
```go
// 時間範囲ベースの固定シード生成
func generateSeed(startTime, endTime time.Time) int64 {
    // 開始時間と終了時間のハッシュ値をシードとして使用
    h := sha256.New()
    h.Write([]byte(startTime.Format(time.RFC3339)))
    h.Write([]byte(endTime.Format(time.RFC3339)))
    hash := h.Sum(nil)
    return int64(binary.BigEndian.Uint64(hash[:8]))
}

func generateLogs(startTime, endTime time.Time) []LogEntry {
    seed := generateSeed(startTime, endTime)
    rng := rand.New(rand.NewSource(seed))
    
    // 同一シードにより、必ず同じログパターンが生成される
    return generateLogsWithRNG(rng, startTime, endTime)
}
```

#### 2. ログID生成の一意性保証
**時間 + 連番による決定論的ID**
```go
func generateLogID(timestamp time.Time, sequence int) string {
    // timestamp（秒単位） + 6桁0埋め連番
    return fmt.Sprintf("log_%d_%06d", timestamp.Unix(), sequence)
}

// 例: "log_1691836800_000001", "log_1691836800_000002"
```

#### 3. 時系列データの整合性保証
**分単位での固定ログ頻度**
```go
func generateLogsForHour(baseTime time.Time, rng *rand.Rand) []LogEntry {
    logs := []LogEntry{}
    
    // 1時間を分単位で分割し、各分のログ数を決定論的に決定
    for minute := 0; minute < 60; minute++ {
        currentTime := baseTime.Add(time.Duration(minute) * time.Minute)
        
        // この分のログ数（0-3件、重み付きランダム）
        logCount := rng.Intn(4) // 0, 1, 2, 3のいずれか
        
        for i := 0; i < logCount; i++ {
            // 分内での秒数をランダムに決定（ただし決定論的）
            seconds := rng.Intn(60)
            timestamp := currentTime.Add(time.Duration(seconds) * time.Second)
            
            log := generateSingleLog(timestamp, rng)
            logs = append(logs, log)
        }
    }
    
    return logs
}
```

#### 4. 異常パターンの決定論的配置
**時間帯別の異常パターン注入**
```go
func injectAnomalousPatterns(logs []LogEntry, timeRange TimeRange, rng *rand.Rand) []LogEntry {
    // 異常パターンの発生条件を決定論的に判定
    anomalyProbability := calculateAnomalyProbability(timeRange)
    
    if rng.Float64() < anomalyProbability {
        // 特定の時間帯に異常パターンを注入
        anomalyType := selectAnomalyType(timeRange, rng)
        anomalousLogs := generateAnomalousLogs(anomalyType, timeRange, rng)
        
        // 決定論的な位置に異常ログを挿入
        insertPosition := rng.Intn(len(logs) + 1)
        logs = insertLogs(logs, anomalousLogs, insertPosition)
    }
    
    return logs
}

func calculateAnomalyProbability(timeRange TimeRange) float64 {
    hour := timeRange.Start.Hour()
    
    // 時間帯別の異常確率（固定値）
    switch {
    case hour >= 9 && hour <= 18:  // 業務時間
        return 0.1  // 10%
    case hour >= 19 || hour <= 8:  // 業務時間外
        return 0.3  // 30%（フェーズ3パターン多発）
    default:
        return 0.2  // 20%
    }
}
```

#### 5. データストレージによる冪等性保証
**DynamoDBでの結果キャッシュ**
```go
type LogCache struct {
    TimeRangeKey string    `dynamodb:"time_range_key"`  // "2024081210-2024081211"
    GeneratedAt  time.Time `dynamodb:"generated_at"`
    LogsJSON     string    `dynamodb:"logs_json"`       // JSONL形式
    TTL          int64     `dynamodb:"ttl"`             // 7日後に期限切れ
}

func getOrGenerateLogs(startTime, endTime time.Time) ([]LogEntry, error) {
    // キャッシュキー生成
    cacheKey := fmt.Sprintf("%s-%s", 
        startTime.Format("2006010215"), 
        endTime.Format("2006010215"))
    
    // DynamoDBからキャッシュ確認
    cached, err := getCachedLogs(cacheKey)
    if err == nil && cached != nil {
        return parseLogsFromJSON(cached.LogsJSON), nil
    }
    
    // キャッシュにない場合は生成
    logs := generateLogs(startTime, endTime)
    
    // DynamoDBにキャッシュ保存
    saveCachedLogs(cacheKey, logs)
    
    return logs, nil
}
```

#### 6. バージョニングによる互換性保証
**設定変更時の影響回避**
```go
const (
    GENERATOR_VERSION = "v1.0.0"
)

func generateSeedWithVersion(startTime, endTime time.Time) int64 {
    h := sha256.New()
    h.Write([]byte(GENERATOR_VERSION))      // バージョン番号も含める
    h.Write([]byte(startTime.Format(time.RFC3339)))
    h.Write([]byte(endTime.Format(time.RFC3339)))
    hash := h.Sum(nil)
    return int64(binary.BigEndian.Uint64(hash[:8]))
}
```

### 冪等性検証方法

#### 1. 単体テスト
```go
func TestIdempotency(t *testing.T) {
    start := time.Date(2024, 8, 12, 10, 0, 0, 0, time.UTC)
    end := time.Date(2024, 8, 12, 11, 0, 0, 0, time.UTC)
    
    // 同じ条件で10回実行
    var results [][]LogEntry
    for i := 0; i < 10; i++ {
        logs := generateLogs(start, end)
        results = append(results, logs)
    }
    
    // 全結果が同一であることを確認
    for i := 1; i < len(results); i++ {
        assert.Equal(t, results[0], results[i], 
            "ログ生成結果が冪等でない")
    }
}
```

#### 2. 統合テスト
```go
func TestAPIIdempotency(t *testing.T) {
    // 同じAPIリクエストを複数回実行
    url := "/logs?start_time=2024-08-12T10:00:00Z&end_time=2024-08-12T11:00:00Z"
    
    var responses []string
    for i := 0; i < 5; i++ {
        resp := callAPI(url)
        responses = append(responses, resp.Body)
    }
    
    // レスポンスボディが完全に同一であることを確認
    for i := 1; i < len(responses); i++ {
        assert.Equal(t, responses[0], responses[i], 
            "API応答が冪等でない")
    }
}
```

### 運用上の注意点

#### 1. キャッシュ戦略
- **DynamoDB TTL**: 7日間でキャッシュ自動削除
- **メモリキャッシュ**: Lambda実行中のメモリ使用量制限
- **キャッシュミス時**: 生成処理のタイムアウト対策

#### 2. 設定変更時の影響
- **バージョニング**: 設定変更時は Generator Version を更新
- **段階的移行**: 新旧バージョンの並行稼働期間設定
- **検証期間**: 新バージョンでの冪等性確認期間

#### 3. デバッグ支援
- **生成ログ**: シード値、使用パラメータをCloudWatch Logsに出力
- **再現手順**: 特定時間範囲のログ生成手順をドキュメント化
- **検証ツール**: ローカル環境での冪等性確認スクリプト提供

