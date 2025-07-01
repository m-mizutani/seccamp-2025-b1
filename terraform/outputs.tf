# Outputs for the Terraform configuration

output "aws_region" {
  description = "The AWS region used"
  value       = var.aws_region
}

output "environment" {
  description = "The environment name"
  value       = var.environment
}

output "project_name" {
  description = "The project name"
  value       = var.project_name
} 