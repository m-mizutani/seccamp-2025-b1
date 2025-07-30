# Importer Lambda 要件

## 概要
importerというLambda関数を実装し、定期的にauditlog APIからログを取得してJSONL形式に変換し、後続のconverter処理のためにS3に保存する機能を提供する。

## 機能要件

### 1. 定期実行
- Amazon EventBridgeを使用して5分間隔でLambda関数を自動実行
- cron式: `rate(5 minutes)`
- 状態を持たないステートレス設計

### 2. ログ取得
- auditlog APIエンドポイントからログデータを取得
- API仕様に基づく適切なパラメータ設定：
  - `startTime`: 現在時刻-7分（RFC3339形式）
  - `endTime`: 現在時刻（RFC3339形式）
  - `offset`: 0から開始、ページング処理で増加
  - `limit`: 100（API最大値）
- 固定幅取得による簡素化：
  - 常に過去7分間のログを取得（5分間隔＋2分バッファ）
  - ステートレス設計により状態管理不要
  - 重複ログは許容（Athena側で対応）

### 3. ページング処理
- auditlog APIは`limit`最大100件、`offset`によるページング
- `metadata.total`で総件数を確認
- 複数回APIを呼び出して全ログを取得：
  ```
  1回目: offset=0, limit=100
  2回目: offset=100, limit=100 (total > 100の場合)
  3回目: offset=200, limit=100 (total > 200の場合)
  ...
  ```

### 4. データ変換・結合
- 各APIレスポンスの`logs`フィールドを抽出
- 複数回の取得結果を時系列順に結合
- 各ログエントリをJSONL（JSON Lines）形式に変換
- 1行につき1つのJSONオブジェクトとして出力

### 5. データ圧縮・保存
- 結合したJSONLデータをgzip形式で圧縮
- S3バケットに保存（後続のconverterが処理可能な場所）
- ファイル名形式: `YYYY/MM/DD/HH/import_YYYYMMDD_HHMMSS.jsonl.gz`
- 例: `2024/08/12/10/import_20240812_102030.jsonl.gz`

### 6. 重複処理方針
- ステートレス設計により状態管理は行わない
- 5分間隔実行で7分間のログを取得するため、2分間の重複が発生
- Athena/分析側で重複ログの除去を実装
- シンプルな設計による運用コスト削減を優先

### 7. エラーハンドリング
- API呼び出し失敗時のリトライ処理（最大3回）
- ネットワークエラー、タイムアウトエラーの適切な処理
- S3アップロード失敗時のリトライ処理
- CloudWatch Logsへの詳細なエラーログ出力

## 非機能要件

### 1. セキュリティ
- IAMロールは必要最低限の権限のみ付与：
  - auditlog APIへのHTTPSアクセス権限
  - 指定S3バケットへの書き込み権限（`s3:PutObject`）
  - CloudWatch Logsへの書き込み権限
- 不要なAWSサービスへのアクセス権限は付与しない

### 2. パフォーマンス
- Lambda関数の適切な設定：
  - メモリ: 512MB（JSON処理とgzip圧縮を考慮）
  - タイムアウト: 5分（ページング処理を考慮）
  - 同時実行数制限: 1（重複実行を防止）

### 3. 信頼性
- 冪等性の保証（同じ時間範囲の重複実行でも安全）
- 失敗時の部分的リトライ（最後に成功したページから再開）
- データ整合性の保証（不完全なファイルは保存しない）

### 4. 運用性
- CloudWatch Logsによる実行ログの記録
- 処理件数、実行時間、エラー情報の出力
- CloudWatch Metricsによる監視（処理件数、エラー率）

## 技術要件

### 1. プログラミング言語
- Go言語で実装
- AWS SDK for Go v2を使用

### 2. インフラストラクチャ
- Terraformを使用してAWSリソースを管理
- 必要なAWSサービス：
  - AWS Lambda
  - Amazon S3（ログ保存用バケット）
  - Amazon EventBridge（定期実行スケジューラ）
  - AWS IAM（権限管理）
  - Amazon CloudWatch（ログ・メトリクス）

### 3. 依存関係
- 既存のauditlog API（terraform/lambda/auditlog）
- 後続のconverter Lambda（terraform/lambda/converter）
- S3バケット構造は既存のconverterと互換性を保つ

## 制約事項

### 1. AWS Lambda制約
- 最大実行時間15分以内での処理完了
- メモリ効率的な処理（大量ログ時のストリーミング処理）

### 2. API制約
- auditlog APIのレート制限対応
- 1回のリクエストで最大100件のログのみ取得可能
- offset + limitによるページング必須

### 3. S3制約
- 単一オブジェクトサイズ制限（5TB）への対応
- 効率的なgzip圧縮による容量最適化

### 4. 運用制約
- 5分間隔実行での処理完了必須
- 前回実行との重複・欠損の防止
- コスト効率的なリソース利用

## 出力仕様

### S3オブジェクト構造
```
s3://bucket-name/
├── 2024/
│   ├── 08/
│   │   ├── 12/
│   │   │   ├── 10/
│   │   │   │   ├── import_20240812_100500.jsonl.gz
│   │   │   │   ├── import_20240812_101000.jsonl.gz
│   │   │   │   └── import_20240812_101500.jsonl.gz
```

### JSONL形式
```jsonl
{"id":"log_1691836800_000001","timestamp":"2024-08-12T10:05:00Z","user":{"email":"teacher@muhaijuku.com","name":"田中太郎","domain":"muhaijuku.com"},"event":{"type":"drive","name":"access","action":"view"},"resource":{"name":"grades/math_test_results.xlsx","id":"1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms","type":"file"},"metadata":{"ip_address":"192.168.1.100","user_agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36","location":{"country":"Japan","region":"Tokyo","city":"Shibuya"}},"result":{"success":true,"denied_reason":null}}
{"id":"log_1691836800_000002","timestamp":"2024-08-12T10:05:30Z","user":{"email":"admin@muhaijuku.com","name":"管理者","domain":"muhaijuku.com"},"event":{"type":"admin","name":"user_create","action":"create"},"resource":{"name":"users/new_teacher","id":"user_12345","type":"user"},"metadata":{"ip_address":"192.168.1.10","user_agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36","location":{"country":"Japan","region":"Tokyo","city":"Shibuya"}},"result":{"success":true,"denied_reason":null}}
```