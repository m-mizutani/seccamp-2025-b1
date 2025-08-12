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

### 3. terraform.tfvars の作成（オプション）

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