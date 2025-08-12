# Terraform Deployment Guide

このディレクトリには、Security Camp 2025 B1コースのクラウドプラットフォームセキュリティ監視基盤を構築するためのTerraformコードが含まれています。

## 前提条件

1. **AWS CLI**: インストールされ、適切な認証情報が設定されていること
2. **Terraform**: バージョン1.0以上
3. **Go**: 1.21以上（Lambda関数のビルドに必要）
4. **GitHub Actions OIDC**: CI/CDパイプライン用（オプション）

## 事前準備

### 1. S3バックエンドバケットの作成

Terraformのstateを管理するためのS3バケットを作成する必要があります：

```bash
aws s3 mb s3://seccamp2025-b1-terraform --region ap-northeast-1
aws s3api put-bucket-versioning \
  --bucket seccamp2025-b1-terraform \
  --versioning-configuration Status=Enabled
```

### 2. teams.json の設定

`teams.json` ファイルは、チームごとのリソースを作成するための設定ファイルです。各チームにGitHubユーザー名を割り当てます：

```json
{
  "teams": {
    "blue": "github-username-1",
    "green": "github-username-2",
    "purple": "github-username-3",
    // ... 他のチーム
    "teal": "",      // 空文字列の場合、このチームのリソースは作成されません
    "gold": "",
    "silver": ""
  }
}
```

**重要**: 
- GitHubユーザー名が設定されているチームのみリソースが作成されます
- 空文字列のチームはスキップされます
- このファイルはgitに登録されているため、実際のユーザー名を設定する場合は注意してください

### 3. AuditLog Lambda用のシードデータの準備

AuditLog Lambdaは、テスト用ログデータを生成するためのシードファイルを必要とします。シードデータは `tools/loggen` を使用して生成します。

#### シードデータの生成とアップロード

1. **loggenツールのビルド**：
   ```bash
   cd tools/loggen
   go build -o loggen
   ```

2. **シードデータの生成**：
   ```bash
   # デフォルト設定でローカルに生成
   ./loggen generate
   
   # 特定の日付でシードを生成
   ./loggen generate --date 2024-08-12
   
   # 異常ログの比率を変更（デフォルト: 15%）
   ./loggen generate --anomaly-ratio 0.20
   ```

3. **S3へ直接アップロード（推奨）**：
   ```bash
   # Terraformデプロイ完了後に実行
   # 圧縮バイナリ形式でS3に直接出力
   ./loggen generate \
     --output s3://seccamp2025-b1-auditlog-seeds/ \
     --format binary-compressed
   ```

   または、既存のファイルをアップロード：
   ```bash
   aws s3 cp ./output/seeds/day_2024-08-12.bin.gz \
     s3://seccamp2025-b1-auditlog-seeds/seeds/large-seed.bin.gz
   ```

#### 生成されるログデータ

loggenは以下の異常パターンを含むログデータを生成します：

**時間帯限定パターン**：
- 夜間の管理者による大量ダウンロード（18:00-9:00）
- 外部リンクアクセスバースト（10:00-16:00）
- VPN経由の水平移動攻撃（9:00-18:00）

**常時発生型パターン**（24時間検知可能）：
- 高頻度認証攻撃（1分に3-5回の認証失敗）
- 超高速データ窃取（1分に10-15件のダウンロード）
- マルチサービス不正アクセス（複数サービスへの探索）
- 地理的同時アクセス（2カ国から同時操作）

**注意**: 
- シードファイルが配置されていない場合、AuditLog Lambdaはエラーを返します
- シードファイルは圧縮形式（.bin.gz）である必要があります
- Lambda側は `seeds/large-seed.bin.gz` というファイル名を期待しています

### 4. terraform.tfvars の作成（オプション）

デフォルト値を変更する場合は、`terraform.tfvars` ファイルを作成します：

```bash
cp terraform.tfvars.example terraform.tfvars
```

設定可能な変数：
- `basename`: リソース名のプレフィックス（デフォルト: "seccamp2025-b1"）
- `aws_region`: AWSリージョン（デフォルト: "ap-northeast-1"）

## デプロイ手順

### 1. 初期化

```bash
cd terraform
terraform init
```

### 2. 設定の検証

```bash
terraform validate
terraform plan
```

### 3. デプロイ

```bash
terraform apply
```

確認プロンプトで `yes` を入力してデプロイを実行します。

### 4. Lambda関数の更新

Lambda関数のコードを変更した場合、Terraformが自動的に再ビルドとデプロイを行います：

```bash
# コード変更後
terraform apply
```

### 5. シードデータのアップロード（初回のみ）

デプロイ完了後、AuditLog Lambda用のシードデータをS3にアップロードします：

```bash
# tools/loggenを使用して直接S3にアップロード（推奨）
cd tools/loggen
./loggen generate \
  --output s3://seccamp2025-b1-auditlog-seeds/ \
  --format binary-compressed

# または、既存のシードファイルをアップロード
aws s3 cp ./output/seeds/day_2024-08-12.bin.gz \
  s3://seccamp2025-b1-auditlog-seeds/seeds/large-seed.bin.gz
```

これにより、AuditLog Lambdaがテストログを生成できるようになります。

## zenv を使用する場合

`zenv` ツールが利用可能な場合、以下のように実行できます：

```bash
cd terraform
zenv terraform init
zenv terraform plan
zenv terraform apply
```

**注意**: `zenv exec` や `zenv exec --` の形式は使用しないでください。

## デプロイされるリソース

### 共有リソース
- **Security Lake**: セキュリティログの中央リポジトリ
- **S3バケット**: 生ログ保存用
- **SNS/SQS**: イベント駆動処理用
- **Glue Crawler**: データカタログ作成用
- **Athena**: クエリ実行環境

### Lambda関数
1. **Importer**: 外部APIからログを取得（Google Workspace等）
2. **Converter**: JSONLからOCSF Parquet形式への変換
3. **AuditLog**: テスト用ログ生成

### チーム別リソース
`teams.json` に設定されたチームごとに以下が作成されます：
- 専用のIAMロール
- チーム固有のリソースタグ

### GitHub Actions連携
- **OIDC Provider**: GitHub Actionsからの認証用
- **IAMロール**: 
  - `GitHubActionsRole-MainRepo`: 管理者権限（メインリポジトリ用）
  - `GitHubActionsRole-B1Secmon`: Lambda管理権限（b1-secmonリポジトリ用）


## トラブルシューティング

### state lockエラー
```bash
terraform force-unlock <LOCK_ID>
```

### Lambda関数のビルドエラー
Lambda関数は自動的にビルドされますが、手動でビルドする場合：

```bash
# Converter
cd lambda/converter
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap main.go types.go s3_interface.go convert.go

# Importer
cd lambda/importer
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap .

# AuditLog
cd lambda/auditlog
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap main.go
```

### リソースの削除
```bash
terraform destroy
```

**警告**: すべてのリソースが削除されます。本番環境では十分注意してください。

## 参考情報

- Lambda関数の詳細は各ディレクトリの `README.md` または `spec.md` を参照
- OCSF形式の詳細は `lambda/converter/schema.md` を参照
- プロジェクト全体のガイドラインは `/CLAUDE.md` を参照