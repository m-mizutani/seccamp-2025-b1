# Main Terraform configuration for SecCamp 2025 B-1

# Data source to get current AWS account ID
data "aws_caller_identity" "current" {}

locals {
  account_id = data.aws_caller_identity.current.account_id
  common_tags = {
    Project   = "seccamp-2025-b1"
    ManagedBy = "terraform"
  }
}

# Team module is removed - all resources are now shared
