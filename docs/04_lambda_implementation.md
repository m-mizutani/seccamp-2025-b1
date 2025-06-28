# Lambda関数（Go）の基本実装とCI/CDセットアップ

## 🚀 事前準備済み環境の説明とセットアップ

### GitHubリポジトリの構成

#### プロジェクト構造
```
seccamp-2025-detection/
├── cmd/
│   └── lambda/
│       └── main.go                # Lambda エントリーポイント
├── internal/
│   ├── detector/
│   │   ├── network.go            # ネットワーク異常検知
│   │   ├── dns.go                # DNS異常検知  
│   │   └── privilege.go          # 特権エスカレーション検知
│   ├── athena/
│   │   └── client.go             # Athenaクライアント
│   └── alert/
│       └── sns.go                # SNS通知
├── sql/
│   ├── network_anomaly.sql       # ネットワーク検知クエリ
│   ├── dns_anomaly.sql           # DNS検知クエリ
│   └── privilege_escalation.sql  # 特権エスカレーション検知クエリ
├── .github/workflows/
│   └── deploy.yml                # GitHub Actions
├── terraform/
│   ├── main.tf                   # インフラ定義
│   └── variables.tf
├── go.mod
├── go.sum
└── README.md
```

### 環境変数設定

#### Lambda環境変数
```bash
# Security Lake設定
ATHENA_WORKGROUP=security-lake-workgroup
ATHENA_OUTPUT_BUCKET=aws-security-lake-query-results-123456789012
SECURITY_LAKE_DATABASE=amazon_security_lake_glue_db_us_east_1

# 通知設定
SNS_TOPIC_ARN=arn:aws:sns:us-east-1:123456789012:security-alerts
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/T.../B.../...

# 検知設定
DETECTION_SCENARIO=network  # network / dns / privilege
```

## 💻 Lambda関数テンプレートの理解

### メイン関数
```go
// cmd/lambda/main.go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"

    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/athena"
    "github.com/aws/aws-sdk-go-v2/service/sns"
    
    "seccamp-detection/internal/detector"
    "seccamp-detection/internal/alert"
)

type SecurityAlert struct {
    Severity    string                 `json:"severity"`
    Title       string                 `json:"title"`
    Description string                 `json:"description"`
    SourceData  string                 `json:"source_data"`
    Timestamp   string                 `json:"timestamp"`
    Details     map[string]interface{} `json:"details"`
}

type LambdaEvent struct {
    Scenario string `json:"scenario"` // "network", "dns", "privilege"
}

func handler(ctx context.Context, event LambdaEvent) error {
    // AWS SDKクライアント初期化
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return fmt.Errorf("failed to load AWS config: %w", err)
    }

    athenaClient := athena.NewFromConfig(cfg)
    snsClient := sns.NewFromConfig(cfg)
    
    // 検知シナリオの選択
    scenario := event.Scenario
    if scenario == "" {
        scenario = os.Getenv("DETECTION_SCENARIO")
    }

    // 検知実行
    alerts, err := executeDetection(ctx, athenaClient, scenario)
    if err != nil {
        log.Printf("Detection failed: %v", err)
        return err
    }

    // アラート送信
    for _, alert := range alerts {
        if err := sendAlert(ctx, snsClient, alert); err != nil {
            log.Printf("Failed to send alert: %v", err)
        }
    }

    log.Printf("Processed %d alerts for scenario: %s", len(alerts), scenario)
    return nil
}

func executeDetection(ctx context.Context, client *athena.Client, scenario string) ([]SecurityAlert, error) {
    switch scenario {
    case "network":
        return detector.DetectNetworkAnomalies(ctx, client)
    case "dns":
        return detector.DetectDNSAnomalies(ctx, client)
    case "privilege":
        return detector.DetectPrivilegeEscalation(ctx, client)
    default:
        return nil, fmt.Errorf("unknown scenario: %s", scenario)
    }
}

func sendAlert(ctx context.Context, client *sns.Client, alert SecurityAlert) error {
    return alert.PublishToSNS(ctx, client, os.Getenv("SNS_TOPIC_ARN"))
}

func main() {
    lambda.Start(handler)
}
```

### Athenaクライアント実装
```go
// internal/athena/client.go
package athena

import (
    "context"
    "fmt"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/athena"
    "github.com/aws/aws-sdk-go-v2/service/athena/types"
)

type Client struct {
    athena    *athena.Client
    workgroup string
    database  string
}

func NewClient(athenaClient *athena.Client, workgroup, database string) *Client {
    return &Client{
        athena:    athenaClient,
        workgroup: workgroup,
        database:  database,
    }
}

func (c *Client) ExecuteQuery(ctx context.Context, query string) (*athena.GetQueryResultsOutput, error) {
    // クエリ実行
    startOutput, err := c.athena.StartQueryExecution(ctx, &athena.StartQueryExecutionInput{
        QueryString: aws.String(query),
        WorkGroup:   aws.String(c.workgroup),
        QueryExecutionContext: &types.QueryExecutionContext{
            Database: aws.String(c.database),
        },
    })
    if err != nil {
        return nil, fmt.Errorf("failed to start query: %w", err)
    }

    // クエリ完了まで待機
    executionId := startOutput.QueryExecutionId
    for {
        execOutput, err := c.athena.GetQueryExecution(ctx, &athena.GetQueryExecutionInput{
            QueryExecutionId: executionId,
        })
        if err != nil {
            return nil, fmt.Errorf("failed to get query execution: %w", err)
        }

        state := execOutput.QueryExecution.Status.State
        switch state {
        case types.QueryExecutionStateSucceeded:
            // 結果取得
            return c.athena.GetQueryResults(ctx, &athena.GetQueryResultsInput{
                QueryExecutionId: executionId,
            })
        case types.QueryExecutionStateFailed, types.QueryExecutionStateCancelled:
            return nil, fmt.Errorf("query failed with state: %s", state)
        default:
            // 実行中、少し待機
            time.Sleep(2 * time.Second)
        }
    }
}
```

### アラート通知実装
```go
// internal/alert/sns.go
package alert

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/sns"
)

type SecurityAlert struct {
    Severity    string                 `json:"severity"`
    Title       string                 `json:"title"`
    Description string                 `json:"description"`
    SourceData  string                 `json:"source_data"`
    Timestamp   string                 `json:"timestamp"`
    Details     map[string]interface{} `json:"details"`
}

func (a SecurityAlert) PublishToSNS(ctx context.Context, client *sns.Client, topicArn string) error {
    // Slack用メッセージフォーマット
    slackMessage := map[string]interface{}{
        "text": fmt.Sprintf("🚨 Security Alert: %s", a.Title),
        "attachments": []map[string]interface{}{
            {
                "color": getSeverityColor(a.Severity),
                "fields": []map[string]interface{}{
                    {"title": "Severity", "value": a.Severity, "short": true},
                    {"title": "Source", "value": a.SourceData, "short": true},
                    {"title": "Time", "value": a.Timestamp, "short": true},
                    {"title": "Description", "value": a.Description, "short": false},
                },
            },
        },
    }

    message, err := json.Marshal(slackMessage)
    if err != nil {
        return fmt.Errorf("failed to marshal slack message: %w", err)
    }

    _, err = client.Publish(ctx, &sns.PublishInput{
        TopicArn: aws.String(topicArn),
        Message:  aws.String(string(message)),
        Subject:  aws.String(a.Title),
    })
    
    return err
}

func getSeverityColor(severity string) string {
    switch severity {
    case "HIGH":
        return "danger"
    case "MEDIUM":
        return "warning"
    default:
        return "good"
    }
}
```

## ⚙️ GitHub Actionによる自動デプロイ

### CI/CDパイプライン設定
```yaml
# .github/workflows/deploy.yml
name: Deploy Security Detection Lambda

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

env:
  AWS_REGION: us-east-1
  GO_VERSION: 1.21

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}
    
    - name: Test
      run: |
        go mod download
        go test -v ./...
    
    - name: Lint
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest

  build-and-deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    permissions:
      id-token: write
      contents: read
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Configure AWS credentials
      uses: aws-actions/configure-aws-credentials@v4
      with:
        role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
        aws-region: ${{ env.AWS_REGION }}
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}
    
    - name: Build Lambda
      run: |
        GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bootstrap cmd/lambda/main.go
        zip lambda-deployment.zip bootstrap
    
    - name: Deploy Lambda
      run: |
        aws lambda update-function-code \
          --function-name security-detection-lambda \
          --zip-file fileb://lambda-deployment.zip
        
        aws lambda wait function-updated \
          --function-name security-detection-lambda
    
    - name: Update Environment Variables
      run: |
        aws lambda update-function-configuration \
          --function-name security-detection-lambda \
          --environment Variables="{
            ATHENA_WORKGROUP=${{ secrets.ATHENA_WORKGROUP }},
            ATHENA_OUTPUT_BUCKET=${{ secrets.ATHENA_OUTPUT_BUCKET }},
            SECURITY_LAKE_DATABASE=${{ secrets.SECURITY_LAKE_DATABASE }},
            SNS_TOPIC_ARN=${{ secrets.SNS_TOPIC_ARN }},
            DETECTION_SCENARIO=${{ github.event.inputs.scenario || 'network' }}
          }"
```

### デプロイ確認とログ確認

#### デプロイ状況確認
```bash
# GitHub Actions実行状況確認
gh run list

# 最新デプロイの詳細確認
gh run view --log
```

#### Lambda関数確認
```bash
# 関数情報確認
aws lambda get-function --function-name security-detection-lambda

# 環境変数確認
aws lambda get-function-configuration --function-name security-detection-lambda
```

#### ログ監視
```bash
# CloudWatch Logsでのリアルタイム監視
aws logs tail /aws/lambda/security-detection-lambda --follow

# 特定期間のログ確認
aws logs filter-log-events \
  --log-group-name /aws/lambda/security-detection-lambda \
  --start-time $(date -d '1 hour ago' +%s)000
```

## 📊 通知フローの理解

### Lambda → SNS → 監視ツール → Slack連携

#### 通知アーキテクチャ
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Lambda    │───▶│     SNS     │───▶│ EventBridge │───▶│    Slack    │
│ (検知処理)   │    │ (メッセージ  │    │ (ルーティング)│    │ (通知表示)   │
│             │    │  配信)      │    │             │    │             │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                           │
                           ▼
                   ┌─────────────┐
                   │ CloudWatch  │
                   │ (ログ保存)   │
                   └─────────────┘
```

#### SNS Topic設定例
```bash
# SNSトピック確認
aws sns get-topic-attributes --topic-arn arn:aws:sns:us-east-1:123456789012:security-alerts

# Slack webhook URL設定（SecretsManagerに保存）
aws secretsmanager get-secret-value --secret-id slack-webhook-url
```

### アラート通知の確認方法

#### Slackでの通知例
```json
{
  "text": "🚨 Security Alert: Network Anomaly Detected",
  "attachments": [
    {
      "color": "danger",
      "fields": [
        {"title": "Severity", "value": "HIGH", "short": true},
        {"title": "Source", "value": "VPC Flow Logs", "short": true},
        {"title": "Time", "value": "2024-08-12T10:30:00Z", "short": true},
        {"title": "Description", "value": "Large data transfer detected from internal network", "short": false}
      ]
    }
  ]
}
```

#### 通知テスト実行
```bash
# Lambda関数の手動実行
aws lambda invoke \
  --function-name security-detection-lambda \
  --payload '{"scenario": "network"}' \
  response.json

# 実行結果確認
cat response.json
```

## 🛠️ 実習での作業手順

### 1. リポジトリクローン
```bash
git clone https://github.com/seccamp-2025/detection-template.git
cd detection-template
```

### 2. Go依存関係確認
```bash
go mod tidy
go test ./...
```

### 3. ローカルテスト実行
```bash
# 環境変数設定
export AWS_PROFILE=seccamp-lab
export DETECTION_SCENARIO=network

# ローカル実行（テスト用）
go run cmd/lambda/main.go
```

### 4. コード修正・デプロイ
```bash
# 変更をコミット
git add .
git commit -m "Add custom detection logic"
git push origin main

# GitHub Actionsでの自動デプロイを確認
gh run watch
```

## 💡 実装のポイント

### 1. エラーハンドリング
```go
// 適切なエラー処理
if err != nil {
    log.Printf("Query execution failed: %v", err)
    // 部分的な失敗でも処理を継続
    continue
}
```

### 2. タイムアウト設定
```go
// コンテキストタイムアウト
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()
```

### 3. ログ出力
```go
// 構造化ログ
log.Printf("Detection completed: scenario=%s, alerts=%d, duration=%v", 
    scenario, len(alerts), time.Since(start))
```

---

**次回**: Security Lake基礎クエリの実践に進みます！ 