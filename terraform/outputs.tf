# Outputs for the Terraform configuration

output "aws_region" {
  description = "The AWS region used"
  value       = var.aws_region
}

output "basename" {
  description = "The base name for resources"
  value       = var.basename
}

output "raw_logs_bucket" {
  description = "Name of the raw logs S3 bucket"
  value       = aws_s3_bucket.raw_logs.bucket
}

output "security_lake_bucket" {
  description = "Security Lake S3 bucket ARN"
  value       = aws_securitylake_data_lake.main.s3_bucket_arn
}

output "auditlog_lambda_url" {
  description = "URL for the AuditLog Lambda Function"
  value       = aws_lambda_function_url.auditlog.function_url
}

# Team-specific outputs
output "teams" {
  description = "Map of team resources"
  value = {
    for team_name, team_module in module.team : team_name => {
      team_name                     = team_module.team_name
      raw_logs_bucket_name          = team_module.raw_logs_bucket_name
      raw_logs_bucket_arn           = team_module.raw_logs_bucket_arn
      raw_logs_sns_topic_arn        = team_module.raw_logs_sns_topic_arn
      raw_logs_sqs_queue_arn        = team_module.raw_logs_sqs_queue_arn
      converter_lambda_function_arn = team_module.converter_lambda_function_arn
      converter_iam_role_name       = team_module.converter_iam_role_name
      importer_iam_role_name        = team_module.importer_iam_role_name
      detector_iam_role_name        = team_module.detector_iam_role_name
      custom_log_source_name        = team_module.custom_log_source_name
    }
  }
} 