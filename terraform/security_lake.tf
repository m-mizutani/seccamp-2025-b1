###########################################
# Amazon Security Lake Configuration
###########################################

# KMS Key for Security Lake encryption - REMOVED
# Using S3_MANAGED_KEY instead to avoid complexity

# Data source for existing V2 role
data "aws_iam_role" "security_lake_meta_store_v2" {
  name = "AmazonSecurityLakeMetaStoreManagerV2"
}

# Security Lake Data Lake
resource "aws_securitylake_data_lake" "main" {
  meta_store_manager_role_arn = data.aws_iam_role.security_lake_meta_store_v2.arn

  configuration {
    region = var.aws_region
    # Use S3 managed encryption instead of KMS to avoid complexity
    encryption_configuration {
      kms_key_id = "S3_MANAGED_KEY"
    }
    lifecycle_configuration {
      expiration {
        days = 365
      }
      transition {
        days          = 30
        storage_class = "STANDARD_IA"
      }
      transition {
        days          = 90
        storage_class = "GLACIER"
      }
    }
  }

  tags = merge(local.common_tags, {
    Name = "${var.basename}-security-lake"
    Type = "security-lake"
  })
}

# Note: Using existing AWS-managed service role AmazonSecurityLakeMetaStoreManagerV2
# This role is automatically created by AWS Security Lake when custom log sources are configured

###########################################
# Custom Log Source (for application logs)
###########################################

# Custom source for Google Workspace logs
resource "aws_securitylake_custom_log_source" "google_workspace" {
  source_name    = "google-workspace"
  source_version = "1.0"

  event_classes = ["API_ACTIVITY", "FILE_ACTIVITY", "AUTHENTICATION", "AUTHORIZATION"]

  configuration {
    crawler_configuration {
      role_arn = aws_iam_role.security_lake_crawler.arn
    }
    provider_identity {
      external_id = "custom-google-workspace-${random_id.external_id.hex}"
      principal   = data.aws_caller_identity.current.account_id
    }
  }

  depends_on = [aws_securitylake_data_lake.main]
}

# Note: Security Lake automatically creates and manages a Glue Crawler
# when a custom log source is created. The crawler name will be the same
# as the source_name (e.g., "google-workspace")

# Random ID for external ID
resource "random_id" "external_id" {
  byte_length = 8
}

# IAM Role for Security Lake Custom Log Source Crawler
resource "aws_iam_role" "security_lake_crawler" {
  name = "${var.basename}-security-lake-crawler-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = [
            "securitylake.amazonaws.com",
            "glue.amazonaws.com"
          ]
        }
      }
    ]
  })

  tags = merge(local.common_tags, {
    Name = "${var.basename}-security-lake-crawler-role"
    Type = "iam-role"
  })
}

# Policy for Security Lake Crawler to access Security Lake S3 bucket
resource "aws_iam_policy" "security_lake_crawler" {
  name        = "${var.basename}-security-lake-crawler-policy"
  description = "Policy for Security Lake Crawler to access Security Lake S3 bucket"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket",
          "s3:GetBucketLocation"
        ]
        Resource = [
          aws_securitylake_data_lake.main.s3_bucket_arn,
          "${aws_securitylake_data_lake.main.s3_bucket_arn}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "glue:GetDatabase",
          "glue:GetTable",
          "glue:GetPartitions",
          "glue:CreateTable",
          "glue:UpdateTable",
          "glue:BatchCreatePartition",
          "glue:BatchUpdatePartition",
          "glue:GetDatabases",
          "glue:GetTables",
          "glue:CreateDatabase",
          "glue:UpdateDatabase"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "lakeformation:GetDataAccess",
          "lakeformation:GrantPermissions",
          "lakeformation:RevokePermissions",
          "lakeformation:BatchGrantPermissions",
          "lakeformation:BatchRevokePermissions",
          "lakeformation:ListPermissions"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "security_lake_crawler" {
  policy_arn = aws_iam_policy.security_lake_crawler.arn
  role       = aws_iam_role.security_lake_crawler.name
}

# Attach AWS Glue Service Role policy
resource "aws_iam_role_policy_attachment" "security_lake_crawler_glue" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSGlueServiceRole"
  role       = aws_iam_role.security_lake_crawler.name
}

###########################################
# Outputs
###########################################

output "security_lake_data_lake_arn" {
  description = "ARN of the Security Lake Data Lake"
  value       = aws_securitylake_data_lake.main.arn
}

output "security_lake_s3_bucket_name" {
  description = "Name of the Security Lake S3 bucket"
  value       = replace(aws_securitylake_data_lake.main.s3_bucket_arn, "arn:aws:s3:::", "")
}

output "security_lake_glue_database_name" {
  description = "Name of the Security Lake Glue database"
  value       = "amazon_security_lake_glue_db_${replace(var.aws_region, "-", "_")}"
}

output "custom_log_source_name" {
  description = "Name of the custom log source"
  value       = aws_securitylake_custom_log_source.google_workspace.source_name
}

output "security_lake_glue_table_name" {
  description = "Name of the Security Lake Glue table for Google Workspace logs"
  value       = "amazon_security_lake_table_${replace(var.aws_region, "-", "_")}_ext_google_workspace_1_0"
}

###########################################
# Lake Formation Permissions for Detector Lambda
###########################################

# Grant permissions to Detector Lambda role for Security Lake database
resource "aws_lakeformation_permissions" "detector_database" {
  principal = aws_iam_role.detector_lambda.arn

  permissions = ["DESCRIBE"]

  database {
    name = "amazon_security_lake_glue_db_${replace(var.aws_region, "-", "_")}"
  }
}

# Grant permissions to Detector Lambda role for Security Lake tables
resource "aws_lakeformation_permissions" "detector_tables" {
  principal = aws_iam_role.detector_lambda.arn

  permissions = ["SELECT", "DESCRIBE"]

  table {
    database_name = "amazon_security_lake_glue_db_${replace(var.aws_region, "-", "_")}"
    wildcard      = true
  }
} 