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

# Shared resource outputs
output "raw_logs_sns_topic_arn" {
  description = "SNS topic ARN for raw logs"
  value       = aws_sns_topic.raw_logs.arn
}

output "raw_logs_sqs_queue_arn" {
  description = "SQS queue ARN for raw logs"
  value       = aws_sqs_queue.raw_logs.arn
}

output "converter_lambda_function_arn" {
  description = "Converter Lambda function ARN"
  value       = aws_lambda_function.converter.arn
}

output "converter_lambda_function_name" {
  description = "Converter Lambda function name"
  value       = aws_lambda_function.converter.function_name
}

output "alerts_sns_topic_arn" {
  description = "SNS topic ARN for alerts"
  value       = aws_sns_topic.alerts.arn
}

output "athena_results_bucket" {
  description = "S3 bucket for Athena query results"
  value       = aws_s3_bucket.athena_results.bucket
}

output "athena_workgroup" {
  description = "Athena workgroup name"
  value       = aws_athena_workgroup.main.name
}

output "detector_lambda_role_arn" {
  description = "Detector Lambda IAM role ARN"
  value       = aws_iam_role.detector_lambda.arn
}

output "detector_lambda_role_name" {
  description = "Detector Lambda IAM role name"
  value       = aws_iam_role.detector_lambda.name
}

# output "athena_database" {
#   description = "Athena database name"
#   value       = aws_glue_catalog_database.main.name
# }

