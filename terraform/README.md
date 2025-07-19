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