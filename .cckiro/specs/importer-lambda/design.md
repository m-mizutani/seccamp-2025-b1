# Importer Lambda 設計

## 概要

Google Workspace auditlog APIから定期的にログを取得し、JSONL形式に変換してS3に保存するLambda関数の設計。EventBridgeによる5分間隔実行でステートレスな処理を行う。

## アーキテクチャ設計

### システム構成図

```
EventBridge (5分間隔)
       ↓
  Importer Lambda
       ↓ (HTTP GET)
  Auditlog API
       ↓
  Importer Lambda
       ↓ (S3 Put)
  S3 Raw Logs Bucket
       ↓ (S3 Event)
  SNS Notification
       ↓
  Converter Lambda
```

### データフロー

1. **EventBridge Trigger**: 5分間隔でImporter Lambdaを起動
2. **Time Range Calculation**: 現在時刻-7分〜現在時刻の範囲を計算
3. **API Pagination**: offset/limitでauditlog APIから全ログを取得
4. **Data Transformation**: レスポンスのlogsフィールドをJSONL形式に変換
5. **Compression**: gzip圧縮を適用
6. **S3 Upload**: raw-logsバケットに保存
7. **SNS Notification**: S3イベントでconverterトリガー

## モジュール設計

### 1. Main Handler

```go
type ImporterHandler struct {
    httpClient   *http.Client
    s3Client     S3API
    auditlogURL  string
    bucketName   string
    timeout      time.Duration
}

func (h *ImporterHandler) Handle(ctx context.Context, event events.EventBridgeEvent) error
```

**責務**: EventBridge イベントを受信し、メイン処理を統括

### 2. Time Range Calculator

```go
type TimeRange struct {
    StartTime time.Time
    EndTime   time.Time
}

func CalculateTimeRange(now time.Time) TimeRange
```

**責務**: 現在時刻から取得範囲（過去7分間）を計算

### 3. Auditlog API Client

```go
type AuditlogClient struct {
    httpClient *http.Client
    baseURL    string
    timeout    time.Duration
}

type LogResponse struct {
    Metadata ResponseMetadata `json:"metadata"`
    Logs     []LogEntry       `json:"logs"`
}

func (c *AuditlogClient) FetchLogs(ctx context.Context, timeRange TimeRange, offset, limit int) (*LogResponse, error)
func (c *AuditlogClient) FetchAllLogs(ctx context.Context, timeRange TimeRange) ([]LogEntry, error)
```

**責務**: auditlog APIとの通信、ページング処理

### 4. Data Transformer

```go
type JSONLTransformer struct{}

func (t *JSONLTransformer) ToJSONL(logs []LogEntry) ([]byte, error)
func (t *JSONLTransformer) Compress(data []byte) ([]byte, error)
```

**責務**: ログデータのJSONL変換とgzip圧縮

### 5. S3 Uploader

```go
type S3Uploader struct {
    client     S3API
    bucketName string
    region     string
}

func (u *S3Uploader) Upload(ctx context.Context, key string, data []byte) error
func (u *S3Uploader) GenerateKey(timestamp time.Time) string
```

**責務**: S3への圧縮ファイルアップロード、キー生成

## 詳細設計

### API通信仕様

**リクエスト形式**:
```
GET {auditlog_url}/logs?startTime=2024-08-12T10:00:00Z&endTime=2024-08-12T10:07:00Z&offset=0&limit=100
```

**ページング処理**:
```go
func (c *AuditlogClient) FetchAllLogs(ctx context.Context, timeRange TimeRange) ([]LogEntry, error) {
    var allLogs []LogEntry
    offset := 0
    limit := 100
    
    for {
        resp, err := c.FetchLogs(ctx, timeRange, offset, limit)
        if err != nil {
            return nil, err
        }
        
        allLogs = append(allLogs, resp.Logs...)
        
        // 全ログ取得完了判定
        if offset + len(resp.Logs) >= resp.Metadata.Total {
            break
        }
        
        offset += limit
    }
    
    return allLogs, nil
}
```

### S3キー設計

**パス構造**:
```
{bucket}/YYYY/MM/DD/HH/import_YYYYMMDD_HHMMSS.jsonl.gz
```

**例**:
```
seccamp-raw-logs/2024/08/12/10/import_20240812_102030.jsonl.gz
```

**キー生成ロジック**:
```go
func (u *S3Uploader) GenerateKey(timestamp time.Time) string {
    return fmt.Sprintf("%04d/%02d/%02d/%02d/import_%04d%02d%02d_%02d%02d%02d.jsonl.gz",
        timestamp.Year(), timestamp.Month(), timestamp.Day(), timestamp.Hour(),
        timestamp.Year(), timestamp.Month(), timestamp.Day(),
        timestamp.Hour(), timestamp.Minute(), timestamp.Second())
}
```

### エラーハンドリング設計

**リトライ戦略**:
```go
type RetryConfig struct {
    MaxRetries    int
    InitialDelay  time.Duration
    MaxDelay      time.Duration
    BackoffFactor float64
}

func WithRetry(ctx context.Context, config RetryConfig, fn func() error) error
```

**エラー分類**:
- **一時的エラー**: ネットワークエラー、API一時停止 → リトライ
- **永続的エラー**: 認証エラー、不正なパラメータ → 即座に失敗
- **部分的エラー**: 一部ページ取得失敗 → 取得済み分は保存、失敗をログ出力

### 設定管理

**環境変数**:
```go
type Config struct {
    AuditlogURL        string // AUDITLOG_URL
    S3BucketName      string // S3_BUCKET_NAME  
    AWSRegion         string // AWS_REGION
    TimeoutSeconds    int    // TIMEOUT_SECONDS (default: 240)
    MaxRetries        int    // MAX_RETRIES (default: 3)
    BufferMinutes     int    // BUFFER_MINUTES (default: 2)
}

func LoadConfig() (*Config, error)
```

## セキュリティ設計

### IAMロール権限

**最小権限設計**:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:PutObject"
            ],
            "Resource": "arn:aws:s3:::${bucket_name}/*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream", 
                "logs:PutLogEvents"
            ],
            "Resource": "arn:aws:logs:*:*:*"
        }
    ]
}
```

### 通信セキュリティ

- **HTTPS通信**: auditlog APIとの通信はHTTPS必須
- **TLS検証**: 証明書検証を有効化
- **タイムアウト**: 長時間接続を防ぐためのタイムアウト設定

## パフォーマンス設計

### Lambda設定

```hcl
resource "aws_lambda_function" "importer" {
  function_name = "${var.basename}-importer"
  runtime       = "provided.al2"
  architecture  = "arm64"
  memory_size   = 512
  timeout       = 300  # 5分
  
  reserved_concurrent_executions = 1  # 重複実行防止
}
```

### メモリ使用量最適化

- **ストリーミング処理**: 大量ログ時のメモリ効率化
- **早期解放**: 不要なオブジェクトの早期ガベージコレクション
- **バッファサイズ**: 適切なバッファサイズでI/O効率化

### 実行時間最適化

- **並列処理**: 複数ページの並列取得（ただしAPI制限を考慮）
- **Keep-Alive**: HTTP接続の再利用
- **圧縮効率**: gzipレベルの最適化

## 監視・ログ設計

### CloudWatch Logs

**ログレベル**:
- **INFO**: 正常処理の進捗情報
- **WARN**: リトライ可能なエラー
- **ERROR**: 処理失敗、要調査

**出力内容**:
```go
type LogEntry struct {
    Level       string    `json:"level"`
    Message     string    `json:"message"`
    Timestamp   time.Time `json:"timestamp"`
    TimeRange   string    `json:"time_range,omitempty"`
    LogCount    int       `json:"log_count,omitempty"`
    S3Key       string    `json:"s3_key,omitempty"`
    Error       string    `json:"error,omitempty"`
    Duration    string    `json:"duration,omitempty"`
}
```

### CloudWatch Metrics

**カスタムメトリクス**:
- `ImportedLogCount`: 取得ログ数
- `ProcessingDuration`: 処理時間
- `APICallCount`: API呼び出し回数
- `ErrorCount`: エラー発生回数

## 運用設計

### デプロイメント

**Terraform管理リソース**:
- Lambda Function
- IAM Role & Policy
- EventBridge Rule
- CloudWatch Log Group

**CI/CD連携**:
- Goバイナリのビルド
- ZIP圧縮とS3アップロード
- Terraformによるデプロイ

### 障害対応

**想定される障害**:
1. **auditlog API停止**: アラート通知、手動実行による復旧
2. **S3アクセス不可**: IAM権限確認、バケットポリシー確認
3. **Lambda実行時間超過**: メモリ増量、タイムアウト延長
4. **重複実行**: 同時実行数制限の確認

**復旧手順**:
1. CloudWatch Logsでエラー詳細確認
2. 失敗時間帯の特定
3. 手動でのLambda再実行
4. データ整合性の確認

## テスト設計

### 単体テスト

- **TimeRange計算**: 境界値テスト
- **JSONL変換**: フォーマット検証
- **S3キー生成**: パス形式検証
- **エラーハンドリング**: 例外ケーステスト

### 結合テスト

- **auditlog API連携**: モックAPIでの疎通確認
- **S3アップロード**: 実際のS3での動作確認
- **EventBridge連携**: 実際のスケジュール実行

### 負荷テスト

- **大量ログ処理**: 1000件以上のログでの動作確認
- **長時間実行**: タイムアウト限界での処理確認
- **API制限**: レート制限時の動作確認