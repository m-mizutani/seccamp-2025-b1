output "team_name" {
  description = "Name of the team"
  value       = var.team_name
}

output "raw_logs_bucket_name" {
  description = "Name of the raw logs S3 bucket"
  value       = aws_s3_bucket.raw_logs.id
}

output "raw_logs_bucket_arn" {
  description = "ARN of the raw logs S3 bucket"
  value       = aws_s3_bucket.raw_logs.arn
}

output "raw_logs_sns_topic_arn" {
  description = "ARN of the raw logs SNS topic"
  value       = aws_sns_topic.raw_logs.arn
}

output "raw_logs_sqs_queue_arn" {
  description = "ARN of the raw logs SQS queue"
  value       = aws_sqs_queue.raw_logs.arn
}

output "raw_logs_sqs_queue_url" {
  description = "URL of the raw logs SQS queue"
  value       = aws_sqs_queue.raw_logs.url
}

output "converter_lambda_function_name" {
  description = "Name of the converter Lambda function"
  value       = aws_lambda_function.converter.function_name
}

output "converter_lambda_function_arn" {
  description = "ARN of the converter Lambda function"
  value       = aws_lambda_function.converter.arn
}



output "converter_iam_role_name" {
  description = "Name of the converter Lambda IAM role"
  value       = aws_iam_role.converter_lambda.name
}

output "converter_iam_role_arn" {
  description = "ARN of the converter Lambda IAM role"
  value       = aws_iam_role.converter_lambda.arn
}



output "custom_log_source_name" {
  description = "Name of the custom log source"
  value       = aws_securitylake_custom_log_source.team_logs.source_name
}

output "security_lake_crawler_role_arn" {
  description = "ARN of the Security Lake crawler IAM role"
  value       = aws_iam_role.security_lake_crawler.arn
}

 