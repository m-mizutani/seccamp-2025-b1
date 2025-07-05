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