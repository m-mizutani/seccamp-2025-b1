# Terraform Configuration for SecCamp 2025 B-1

このディレクトリには、SecCamp 2025 B-1のためのTerraform設定が含まれています。

## 前提条件

- Terraform >= 1.0がインストールされていること
- AWS CLIが設定されていること
- S3バケット `mztn-seccamp2025-b1-terraform` が存在すること

## 初期設定

1. Terraformを初期化する：
```bash
terraform init
```

2. 設定を検証する：
```bash
terraform validate
```

3. 実行計画を確認する：
```bash
terraform plan
```

4. インフラストラクチャを適用する：
```bash
terraform apply
```

## ファイル構成

- `providers.tf`: Terraformプロバイダーの設定
- `backend.tf`: S3バックエンドの設定
- `variables.tf`: 変数定義
- `main.tf`: メインのリソース定義
- `outputs.tf`: 出力値の定義

## バックエンド設定

このプロジェクトは以下のS3バックエンド設定を使用します：
- バケット: `mztn-seccamp2025-b1-terraform`
- キー: `terraform.tfstate`
- リージョン: `ap-northeast-1`
- 暗号化: 有効

## 変数

主要な変数：
- `aws_region`: AWSリージョン（デフォルト: ap-northeast-1）
- `environment`: 環境名（デフォルト: dev）
- `project_name`: プロジェクト名（デフォルト: seccamp-2025-b1）

## カスタマイズ

必要に応じて `main.tf` にリソースを追加してください。コメントアウトされたサンプルコードを参考にしてください。

## 開発環境向けの設定

開発環境でコストを削減するため、アクティブなリソース（定期実行やパブリックエンドポイント）を無効化できます。

### アクティブリソースの無効化方法

`enable_active_resources` 変数を `false` に設定します：

**方法1: terraform.tfvars を使用**
```bash
# terraform.tfvars を作成または編集
echo "enable_active_resources = false" >> terraform.tfvars
```

**方法2: 環境変数を使用**
```bash
export TF_VAR_enable_active_resources=false
```

**方法3: コマンドラインで指定**
```bash
terraform apply -var="enable_active_resources=false"
```

### アクティブリソースの有効化方法（デフォルト）

アクティブリソースを再度有効にするには、変数を `true` に設定するか、設定を削除します：

**方法1: terraform.tfvars を更新**
```bash
# terraform.tfvars を編集
enable_active_resources = true
```

**方法2: 環境変数を削除**
```bash
unset TF_VAR_enable_active_resources
```

**方法3: デフォルト値を使用**
```bash
# terraform.tfvars から該当行を削除
# デフォルト値は true です
```

### 制御されるリソース

`enable_active_resources = false` の場合、以下のリソースが無効化されます：

1. **auditlog Lambda Function URL**
   - Lambda 関数は残りますが、パブリック URL エンドポイントが削除されます
   - HTTP 経由での監査ログアクセスが無効になります

2. **importer Lambda の定期実行**
   - Lambda 関数は残りますが、EventBridge トリガーが削除されます
   - 5分ごとの自動ログインポートが無効になります
   - 手動実行は引き続き可能です

### 制御できないリソース

**Security Lake Glue Crawler**: AWS Security Lake サービスが自動管理するため、Terraform では制御できません。Glue Crawler を無効化するには：

```bash
# AWS CLI を使用して停止
aws glue stop-crawler --name google-workspace --region ap-northeast-1

# 再度有効化する場合
aws glue start-crawler --name google-workspace --region ap-northeast-1
```

または AWS コンソールから: AWS Glue → Crawlers → google-workspace → Stop/Start

## 注意事項

- `enable_active_resources` のデフォルト値は `true` で、既存環境との互換性を維持します
- 有効/無効を切り替える際、一部のリソースが削除・再作成されます
- 変更を適用する前に必ず `terraform plan` で影響を確認してください