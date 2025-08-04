# Lambda実装と検知ルール作成

**時間：11:00-11:50 (50分)**

## 概要

このパートでは、実際にコードを書いて Security Lake を活用したセキュリティ監視システムの一部を実装します。ログ収集 Lambda の実装と、前のパートで学んだ OCSF スキーマの知識を活かして検知ルールを作成し、自動化された脅威検知システムを構築します。

## 学習目標

- Lambda 関数を使用したログ収集の実装方法を学ぶ
- 外部 API からのデータ取得と S3 への保存処理を理解する
- 実際の脅威シナリオに基づいた検知 SQL を作成する
- セキュリティ監視システムの自動化について実践的に学ぶ

## ログ収集 Lambda 実装（スケルトン式）（25分）

### 環境準備・理解（5分）

#### 1. GitHub リポジトリのクローンとブランチ作成

```bash
# リポジトリのクローン（まだの場合）
git clone https://github.com/m-mizutani/seccamp-2025-b1.git
cd seccamp-2025-b1

# 自分の作業用ブランチを作成
git checkout -b feature/my-implementation-{your-name}
```

#### 2. スケルトンコードの構造確認

```bash
# Lambda 関数のディレクトリ構造を確認
ls -la terraform/lambda/importer/

# スケルトンコードを確認
cat terraform/lambda/importer/main.go
```

主要なファイル：
- `main.go` - メイン処理（実装箇所あり）
- `types.go` - データ型定義
- `go.mod` - 依存関係定義

#### 3. 環境変数とリソースの確認

実装に必要な情報：
- **API_ENDPOINT**: ログ取得元の API URL（環境変数で提供）
- **S3_BUCKET**: ログ保存先の S3 バケット名（環境変数で提供）
- **IAM ロール**: S3 への書き込み権限（設定済み）

### HTTP ログ取得と S3 保存の実装（15分）

#### 1. 実装すべき処理の理解

```go
// terraform/lambda/importer/main.go の実装箇所

func handler(ctx context.Context, event events.CloudWatchEvent) error {
    // 1. 現在時刻を取得し、取得する時間範囲を決定
    //    - 重複を避けつつ、欠損を最小化する時刻調整
    
    // 2. 外部 API からログデータを取得
    //    - HTTP GET リクエストでJSON形式のログを取得
    
    // 3. 取得したデータを JSONL 形式に変換
    //    - 各ログエントリを1行のJSONとして出力
    
    // 4. gzip 圧縮して S3 に保存
    //    - 適切なキー設計（日付ベースのパーティション）
    //    - エラーハンドリング
}
```

#### 2. 時刻調整ロジックの実装

```go
// 実装例：15分前から5分前までのログを取得
func getTimeRange() (start, end time.Time) {
    now := time.Now().UTC()
    
    // 重複回避のため、5分前までのデータを取得
    end = now.Add(-5 * time.Minute)
    
    // 15分前から開始（10分間のウィンドウ）
    start = end.Add(-10 * time.Minute)
    
    return start, end
}
```

#### 3. API からのデータ取得

```go
// HTTP クライアントでログを取得
func fetchLogs(apiEndpoint string, start, end time.Time) ([]LogEntry, error) {
    // クエリパラメータの構築
    params := url.Values{}
    params.Add("start_time", start.Format(time.RFC3339))
    params.Add("end_time", end.Format(time.RFC3339))
    
    // HTTP リクエストの送信
    resp, err := http.Get(fmt.Sprintf("%s?%s", apiEndpoint, params.Encode()))
    if err != nil {
        return nil, fmt.Errorf("failed to fetch logs: %w", err)
    }
    defer resp.Body.Close()
    
    // レスポンスのパース
    var logs []LogEntry
    if err := json.NewDecoder(resp.Body).Decode(&logs); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return logs, nil
}
```

#### 4. JSONL 形式への変換と S3 保存

```go
// JSONL 形式に変換して gzip 圧縮
func saveToS3(logs []LogEntry, bucket string) error {
    // S3 キーの生成（日付ベースのパーティション）
    now := time.Now().UTC()
    key := fmt.Sprintf(
        "raw-logs/year=%d/month=%02d/day=%02d/logs_%s.jsonl.gz",
        now.Year(), now.Month(), now.Day(),
        now.Format("20060102_150405"),
    )
    
    // gzip writer の作成
    var buf bytes.Buffer
    gzWriter := gzip.NewWriter(&buf)
    
    // 各ログエントリを JSONL として書き込み
    for _, log := range logs {
        jsonData, err := json.Marshal(log)
        if err != nil {
            return fmt.Errorf("failed to marshal log: %w", err)
        }
        
        if _, err := gzWriter.Write(jsonData); err != nil {
            return fmt.Errorf("failed to write to gzip: %w", err)
        }
        
        if _, err := gzWriter.Write([]byte("\n")); err != nil {
            return fmt.Errorf("failed to write newline: %w", err)
        }
    }
    
    // gzip writer を閉じる
    if err := gzWriter.Close(); err != nil {
        return fmt.Errorf("failed to close gzip writer: %w", err)
    }
    
    // S3 にアップロード
    _, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(key),
        Body:   bytes.NewReader(buf.Bytes()),
        ContentType: aws.String("application/x-gzip"),
    })
    
    return err
}
```

#### 5. エラーハンドリングとロギング

```go
// 適切なエラーハンドリングとログ出力
func handler(ctx context.Context, event events.CloudWatchEvent) error {
    log.Printf("Starting log import process")
    
    // 環境変数の取得
    apiEndpoint := os.Getenv("API_ENDPOINT")
    s3Bucket := os.Getenv("S3_BUCKET")
    
    if apiEndpoint == "" || s3Bucket == "" {
        return fmt.Errorf("required environment variables not set")
    }
    
    // 時間範囲の決定
    start, end := getTimeRange()
    log.Printf("Fetching logs from %s to %s", start, end)
    
    // ログの取得
    logs, err := fetchLogs(apiEndpoint, start, end)
    if err != nil {
        log.Printf("Error fetching logs: %v", err)
        return err
    }
    
    log.Printf("Fetched %d log entries", len(logs))
    
    // S3 への保存
    if err := saveToS3(logs, s3Bucket); err != nil {
        log.Printf("Error saving to S3: %v", err)
        return err
    }
    
    log.Printf("Successfully saved logs to S3")
    return nil
}
```

## 脅威シナリオ設計と検知 SQL 作成（25分）

### 🏫 無敗塾ベースの脅威シナリオ分析（10分）

#### 実例 1: 夜間の管理者による大量学習データダウンロード

**シナリオ詳細**：
- 通常業務時間外（22時〜6時）のアクセス
- 管理者権限を持つアカウント
- 短時間で大量のファイルダウンロード
- 通常とは異なる IP アドレスからのアクセス

**検知ポイント**：
1. 時間帯の異常性
2. ダウンロード数の異常性
3. アクセス元の異常性
4. 権限レベルとアクティビティの組み合わせ

#### 実例 2: 機密フォルダの意図しない外部流出

**シナリオ詳細**：
- 機密データを含むフォルダへの共有設定変更
- "anyone with link" 権限の設定
- 設定変更後の外部 IP からの大量アクセス
- 短時間での異常なアクセス数増加

**検知ポイント**：
1. 共有設定の変更イベント
2. 外部 IP アドレスからのアクセス
3. アクセス数の急激な増加
4. 機密データマーカーの存在

### SQL 検知クエリの実装（10分）

#### 検知ルール 1: 夜間の大量ダウンロード

```sql
-- 夜間の異常なダウンロード活動を検知
WITH night_downloads AS (
    SELECT 
        actor.user.email_addr as user_email,
        actor.user.type_id as user_type,
        COUNT(DISTINCT web_resources) as download_count,
        COUNT(DISTINCT src_endpoint.ip) as unique_ips,
        ARRAY_JOIN(ARRAY_AGG(DISTINCT web_resources[1].name), ', ') as sample_files,
        MIN(from_unixtime(time/1000)) as first_download,
        MAX(from_unixtime(time/1000)) as last_download,
        DATE_DIFF('minute', 
            MIN(from_unixtime(time/1000)), 
            MAX(from_unixtime(time/1000))
        ) as duration_minutes
    FROM seccamp2025_b1_security_lake.google_workspace
    WHERE activity_id = 7  -- Export/Download
        AND status_id = 1  -- Success
        AND (
            EXTRACT(HOUR FROM from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo') >= 22
            OR EXTRACT(HOUR FROM from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo') < 6
        )
        AND from_unixtime(time/1000) > current_timestamp - interval '24' hour
        AND CARDINALITY(web_resources) > 0
    GROUP BY 
        actor.user.email_addr,
        actor.user.type_id
)
SELECT 
    'NIGHT_MASS_DOWNLOAD' as alert_type,
    'HIGH' as severity,
    user_email,
    CASE 
        WHEN user_type = 2 THEN 'Admin'
        WHEN user_type = 1 THEN 'User'
        ELSE 'Unknown'
    END as user_role,
    download_count,
    unique_ips,
    duration_minutes,
    sample_files,
    first_download,
    last_download,
    CONCAT(
        'User ', user_email, 
        ' downloaded ', CAST(download_count AS VARCHAR), 
        ' files during night hours from ', CAST(unique_ips AS VARCHAR), 
        ' different IPs'
    ) as description
FROM night_downloads
WHERE download_count >= 20  -- 閾値：20ファイル以上
ORDER BY download_count DESC;
```

#### 検知ルール 2: 機密データの外部流出

```sql
-- 共有設定変更後の異常アクセスを検知
WITH sharing_changes AS (
    -- まず共有設定の変更を特定
    SELECT 
        actor.user.email_addr as sharing_user,
        web_resources[1].uid as resource_id,
        web_resources[1].name as resource_name,
        time as share_time
    FROM seccamp2025_b1_security_lake.google_workspace
    WHERE activity_id = 8  -- Share
        AND status_id = 1
        AND CONTAINS(CAST(metadata.labels AS VARCHAR), 'anyone_with_link')
        AND from_unixtime(time/1000) > current_timestamp - interval '24' hour
),
external_access AS (
    -- 外部IPからのアクセスを検出
    SELECT 
        s.resource_id,
        s.resource_name,
        s.sharing_user,
        s.share_time,
        COUNT(DISTINCT a.src_endpoint.ip) as external_ip_count,
        COUNT(*) as access_count,
        ARRAY_JOIN(
            ARRAY_AGG(DISTINCT 
                CASE 
                    WHEN a.src_endpoint.ip NOT LIKE '10.%' 
                    AND a.src_endpoint.ip NOT LIKE '172.16.%' 
                    AND a.src_endpoint.ip NOT LIKE '192.168.%'
                    THEN a.src_endpoint.ip 
                END
            ), ', '
        ) as external_ips
    FROM sharing_changes s
    JOIN seccamp2025_b1_security_lake.google_workspace a
        ON a.web_resources[1].uid = s.resource_id
        AND a.time > s.share_time
        AND a.time < s.share_time + (3600 * 1000)  -- 共有後1時間以内
    WHERE a.activity_id IN (2, 7)  -- Read or Download
        AND a.status_id = 1
        AND (
            a.src_endpoint.ip NOT LIKE '10.%' 
            AND a.src_endpoint.ip NOT LIKE '172.16.%' 
            AND a.src_endpoint.ip NOT LIKE '192.168.%'
        )
    GROUP BY 
        s.resource_id,
        s.resource_name,
        s.sharing_user,
        s.share_time
)
SELECT 
    'EXTERNAL_DATA_LEAK' as alert_type,
    'CRITICAL' as severity,
    sharing_user,
    resource_name,
    from_unixtime(share_time/1000) as share_time,
    external_ip_count,
    access_count,
    external_ips,
    CONCAT(
        'Potential data leak: ', resource_name,
        ' was shared publicly by ', sharing_user,
        ' and accessed ', CAST(access_count AS VARCHAR),
        ' times from ', CAST(external_ip_count AS VARCHAR),
        ' external IPs within 1 hour'
    ) as description
FROM external_access
WHERE access_count >= 10  -- 閾値：10回以上のアクセス
ORDER BY access_count DESC;
```

### Lambda 統合とアラート実装（5分）

#### 検知 Lambda への SQL 組み込み

```go
// terraform/lambda/detection/main.go

type DetectionRule struct {
    Name        string
    SQLQuery    string
    Threshold   int
    Severity    string
}

var detectionRules = []DetectionRule{
    {
        Name:      "NightMassDownload",
        SQLQuery:  nightMassDownloadSQL,  // 上記のSQL
        Threshold: 1,
        Severity:  "HIGH",
    },
    {
        Name:      "ExternalDataLeak",
        SQLQuery:  externalDataLeakSQL,  // 上記のSQL
        Threshold: 1,
        Severity:  "CRITICAL",
    },
}

func runDetection(ctx context.Context) ([]Alert, error) {
    var alerts []Alert
    
    for _, rule := range detectionRules {
        // Athena でクエリ実行
        results, err := executeAthenaQuery(ctx, rule.SQLQuery)
        if err != nil {
            log.Printf("Error executing rule %s: %v", rule.Name, err)
            continue
        }
        
        // 結果を解析してアラート生成
        if len(results) >= rule.Threshold {
            for _, result := range results {
                alert := Alert{
                    RuleName:    rule.Name,
                    Severity:    rule.Severity,
                    Description: result["description"],
                    Details:     result,
                    Timestamp:   time.Now(),
                }
                alerts = append(alerts, alert)
            }
        }
    }
    
    return alerts, nil
}
```

#### SNS 通知メッセージの構造化

```go
// アラートを SNS メッセージに変換
func formatAlertMessage(alerts []Alert) string {
    if len(alerts) == 0 {
        return "No security alerts detected."
    }
    
    var message strings.Builder
    message.WriteString(fmt.Sprintf(
        "🚨 Security Alert: %d issues detected\n\n", 
        len(alerts),
    ))
    
    // 重要度でソート
    sort.Slice(alerts, func(i, j int) bool {
        return getSeverityLevel(alerts[i].Severity) > 
               getSeverityLevel(alerts[j].Severity)
    })
    
    for i, alert := range alerts {
        message.WriteString(fmt.Sprintf(
            "Alert #%d [%s]\n", 
            i+1, 
            alert.Severity,
        ))
        message.WriteString(fmt.Sprintf(
            "Rule: %s\n", 
            alert.RuleName,
        ))
        message.WriteString(fmt.Sprintf(
            "Description: %s\n", 
            alert.Description,
        ))
        message.WriteString(fmt.Sprintf(
            "Time: %s\n\n", 
            alert.Timestamp.Format("2006-01-02 15:04:05 JST"),
        ))
    }
    
    return message.String()
}
```

#### デプロイとテスト

```bash
# Lambda 関数のビルド
cd terraform/lambda/importer
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap main.go

# Terraform でデプロイ
cd ../../..
terraform plan
terraform apply

# GitHub Actions でのデプロイ（推奨）
git add .
git commit -m "Implement log collection and detection rules"
git push origin feature/my-implementation-{your-name}
```

## 実装のベストプラクティス

### 1. エラーハンドリング
- 一時的なエラーにはリトライ処理を実装
- 永続的なエラーは適切にログ出力
- Dead Letter Queue (DLQ) の活用

### 2. パフォーマンス最適化
- バッチ処理による効率化
- 並行処理の適切な利用
- メモリ使用量の最適化

### 3. セキュリティ考慮事項
- 環境変数での機密情報管理
- 最小権限の原則に基づく IAM ロール
- ログに機密情報を出力しない

### 4. 監視とデバッグ
- CloudWatch Logs での詳細なログ出力
- X-Ray によるトレーシング
- メトリクスの収集と可視化

## まとめ

このパートでは：

1. **ログ収集 Lambda の実装**
   - 外部 API からのデータ取得
   - JSONL 形式への変換と圧縮
   - S3 への効率的な保存

2. **検知ルールの作成**
   - 実際の脅威シナリオに基づいた SQL
   - 複雑な条件を組み合わせた検知ロジック
   - 誤検知を減らすための工夫

3. **自動化システムの構築**
   - Lambda による定期実行
   - Athena との連携
   - SNS によるアラート通知

これらの実装を通じて、クラウド環境でのセキュリティ監視システムの構築方法を実践的に学びました。