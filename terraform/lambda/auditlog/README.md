# AuditLog Lambda

Google Workspaceログを配信するLambda API

## 機能

- **ログ生成**: 事前生成されたシードファイルから時間範囲指定でログを生成
- **REST API**: 時間範囲指定でのログ取得
- **ページネーション**: 大量ログの効率的配信
- **時間フィルタリング**: RFC3339形式での時間範囲指定

## API仕様

### GET /logs

指定時間範囲のGoogle Workspaceログを取得

**パラメータ（すべて必須）:**
- `startTime` (query): 開始時刻 (RFC3339形式: YYYY-MM-DDTHH:MM:SSZ)
- `endTime` (query): 終了時刻 (RFC3339形式: YYYY-MM-DDTHH:MM:SSZ)
- `limit` (query, オプション): 取得件数 (1-100, デフォルト: 100)  
- `offset` (query, オプション): オフセット (デフォルト: 0)

**制限事項:**
- `limit` の最大値は100件
- `endTime` は `startTime` より後の時刻である必要あり
- 利用可能な時間範囲: 2024-08-12 00:00:00Z ～ 2024-08-12 23:59:59Z

**レスポンス例:**
```json
{
  "date": "2024-08-12T09:00:00Z to 2024-08-12T10:00:00Z",
  "metadata": {
    "total": 2450,
    "offset": 0,
    "limit": 100,
    "generated": "2024-08-12T10:30:00Z"
  },
  "logs": [
    {
      "kind": "admin#reports#activity",
      "id": {
        "time": "2024-08-12T09:15:30Z",
        "uniqueQualifier": "12345",
        "applicationName": "drive",
        "customerId": "C03az79cb"
      },
      "actor": {
        "email": "teacher01@muhai-academy.com",
        "profileId": "114511147312345678906"
      },
      "ownerDomain": "muhai-academy.com",
      "ipAddress": "192.168.1.100",
      "events": [
        {
          "type": "access",
          "name": "view",
          "parameters": [
            {
              "name": "doc_id",
              "value": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE1001"
            },
            {
              "name": "doc_title", 
              "value": "教材/数学/教科書.pdf"
            }
          ]
        }
      ]
    }
  ]
}
```

## エラーレスポンス

```json
{
  "error": "Bad Request",
  "message": "limit must not exceed 100"
}
```

**主なエラーパターン:**
- `400`: パラメータ不正（必須パラメータ欠如、形式エラー、制限値超過など）
- `500`: サーバー内部エラー

## サンプルcurlコマンド

```bash
# 基本的な使用例 - 1時間分のログ取得
curl "https://your-lambda-url.lambda-url.us-east-1.on.aws/logs?startTime=2024-08-12T09:00:00Z&endTime=2024-08-12T10:00:00Z"

# 午前中のログを50件ずつ取得
curl "https://your-lambda-url.lambda-url.us-east-1.on.aws/logs?startTime=2024-08-12T09:00:00Z&endTime=2024-08-12T12:00:00Z&limit=50"

# ページネーション - 次の50件を取得
curl "https://your-lambda-url.lambda-url.us-east-1.on.aws/logs?startTime=2024-08-12T09:00:00Z&endTime=2024-08-12T12:00:00Z&limit=50&offset=50"

# 特定の時間帯の大量データ取得（最大100件）
curl "https://your-lambda-url.lambda-url.us-east-1.on.aws/logs?startTime=2024-08-12T14:00:00Z&endTime=2024-08-12T15:00:00Z&limit=100"

# レスポンスを整形して表示
curl -s "https://your-lambda-url.lambda-url.us-east-1.on.aws/logs?startTime=2024-08-12T09:00:00Z&endTime=2024-08-12T09:30:00Z&limit=5" | jq '.'

# メタデータのみ確認（ログ数を確認）
curl -s "https://your-lambda-url.lambda-url.us-east-1.on.aws/logs?startTime=2024-08-12T00:00:00Z&endTime=2024-08-12T23:59:59Z&limit=1" | jq '.metadata'

# エラーテスト - limitの上限超過
curl "https://your-lambda-url.lambda-url.us-east-1.on.aws/logs?startTime=2024-08-12T09:00:00Z&endTime=2024-08-12T10:00:00Z&limit=101"

# エラーテスト - 必須パラメータ欠如
curl "https://your-lambda-url.lambda-url.us-east-1.on.aws/logs?startTime=2024-08-12T09:00:00Z"
```

## ビルド・デプロイ

```bash
# 依存関係の取得
go mod tidy

# ビルド
GOOS=linux GOARCH=amd64 go build -o bootstrap main.go

# ZIP作成
zip lambda-deployment.zip bootstrap
```

## Seedデータ管理

### 概要
Lambda関数は埋め込みデータの代わりにS3バケットから大容量のseedデータを読み込みます。これにより、より大きなデータセットの使用と容易な更新が可能になります。

### S3設定
- バケット: `seccamp2025-b1-auditlog-seeds`
- キー: `seeds/large-seed.bin.gz`
- バケット名は環境変数 `SEED_BUCKET_NAME` で設定

### Seedデータの生成とアップロード

`tools/putseed` ユーティリティを使用してseedデータを生成・アップロード：

```bash
# 10倍のseedデータを生成してアップロード（デフォルト）
cd tools/putseed
go run main.go

# カスタム倍率で生成
go run main.go -multiplier=20

# 既存のseedファイルをベースに使用
go run main.go -existing=../../terraform/lambda/auditlog/seeds/day_2024-08-12.bin.gz

# アップロードせずに生成のみ
go run main.go -upload=false
```

### パフォーマンス最適化
- 初回ダウンロード後、seedデータはメモリにキャッシュ
- 後続の呼び出し（warm start）はキャッシュを使用
- Lambdaコンテナのリサイクル時にキャッシュはクリア

### 必要なIAM権限
- seedデータバケットへの `s3:GetObject`
- seedデータバケットへの `s3:ListBucket`
- 基本的なLambda実行権限