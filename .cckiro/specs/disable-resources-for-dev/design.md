# 開発環境向けリソース無効化機能の設計

## 設計概要

Terraformの `count` メタ引数を使用して、特定のリソースの作成を条件付きで制御する設計とします。

## 詳細設計

### 1. 変数定義

**ファイル**: `terraform/variables.tf`

```hcl
variable "enable_active_resources" {
  description = "Enable active resources like Lambda URLs and scheduled executions"
  type        = bool
  default     = true
}
```

### 2. リソースの条件付き作成

#### 2.1 auditlog Lambda Function URL

**ファイル**: `terraform/lambda.tf`

```hcl
resource "aws_lambda_function_url" "auditlog" {
  count = var.enable_active_resources ? 1 : 0
  
  function_name      = aws_lambda_function.auditlog.function_name
  authorization_type = "NONE"
  
  cors {
    allow_credentials = true
    allow_origins     = ["*"]
    allow_methods     = ["*"]
    allow_headers     = ["date", "keep-alive"]
    expose_headers    = ["keep-alive", "date"]
    max_age           = 86400
  }
}
```

#### 2.2 Importer定期実行

**ファイル**: `terraform/lambda.tf`

EventBridge関連リソースに `count` を追加：

```hcl
resource "aws_cloudwatch_event_rule" "importer_schedule" {
  count = var.enable_active_resources ? 1 : 0
  
  name                = "${var.basename}-importer-schedule"
  description         = "Trigger importer Lambda every hour"
  schedule_expression = "rate(1 hour)"
}

resource "aws_cloudwatch_event_target" "importer_target" {
  count = var.enable_active_resources ? 1 : 0
  
  rule      = aws_cloudwatch_event_rule.importer_schedule[0].name
  target_id = "${var.basename}-importer-target"
  arn       = aws_lambda_function.importer.arn
}

resource "aws_lambda_permission" "allow_eventbridge" {
  count = var.enable_active_resources ? 1 : 0
  
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.importer.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.importer_schedule[0].arn
}
```

#### 2.3 Glue Crawler定期実行

**注意**: Security Lake の Glue Crawler は AWS Security Lake サービスによって自動的に作成・管理されるため、Terraform で直接制御することはできません。`aws_securitylake_custom_log_source` リソースを作成すると、同名の Glue Crawler が自動的に作成され、定期的に実行されます。

この Crawler の実行を無効化したい場合は、AWS コンソールまたは AWS CLI から手動で設定する必要があります。

### 3. Output の更新

**ファイル**: `terraform/outputs.tf`

リソースが条件付きになるため、outputも更新が必要：

```hcl
output "auditlog_lambda_url" {
  description = "URL of the auditlog Lambda function"
  value       = var.enable_active_resources && length(aws_lambda_function_url.auditlog) > 0 ? aws_lambda_function_url.auditlog[0].function_url : "Disabled in development mode"
}

output "active_resources_enabled" {
  description = "Whether active resources are enabled"
  value       = var.enable_active_resources
}
```

### 4. terraform.tfvars の例

**ファイル**: `terraform/terraform.tfvars.example`

```hcl
# Development environment settings
enable_active_resources = false

# Production environment settings (default)
# enable_active_resources = true
```

### 5. 依存関係の処理

`count` を使用する場合、リソースの参照方法が変わるため、以下の点に注意：

- `aws_lambda_function_url.auditlog` → `aws_lambda_function_url.auditlog[0]`
- 条件付きリソースを参照する際は、`try()` 関数や `length()` チェックを使用

## 実装時の注意事項

1. **既存リソースへの影響**
   - 初回適用時は、既存のリソースが一度削除されて再作成される可能性がある
   - 本番環境では慎重にplanを確認してから適用する

2. **エラーハンドリング**
   - 条件付きリソースを参照する箇所では、リソースが存在しない場合のエラーを回避する

3. **ドキュメント更新**
   - README.mdに開発環境での使用方法を追記
   - 変数の説明を明確にする

## 将来の拡張性

この設計により、将来的に他のリソースも同様の方法で制御可能：
- 他のLambda関数のトリガー
- CloudWatch Logsの保持期間
- S3バケットのライフサイクルポリシー

変数名を `enable_active_resources` としているため、アクティブに動作する他のリソースも同じ変数で制御できます。