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

###########################################
# Team Modules
###########################################

module "team" {
  source = "./modules/team"

  for_each = toset(var.teams)

  team_name                 = each.value
  basename                  = var.basename
  aws_region                = var.aws_region
  common_tags               = local.common_tags
  security_lake_bucket      = aws_securitylake_data_lake.main.s3_bucket_arn
  security_lake_database    = "amazon_security_lake_glue_db_${replace(var.aws_region, "-", "_")}"
  alerts_sns_topic_arn      = aws_sns_topic.alerts.arn
  athena_results_bucket     = aws_s3_bucket.athena_results.id
  converter_zip_path        = data.archive_file.converter_lambda_zip.output_path
  converter_zip_hash        = data.archive_file.converter_lambda_zip.output_base64sha256

  depends_on = [
    aws_securitylake_data_lake.main,
    null_resource.build_converter,
  ]
}
