###########################################
# Security Lake Configuration
###########################################

# Security Lake Data Lake
resource "aws_securitylake_data_lake" "main" {
  configuration {
    region = var.aws_region
  }

  tags = merge(local.common_tags, {
    Name = "${var.basename}-security-lake"
  })
}

# Security Lake AWS Log Source for CloudTrail
resource "aws_securitylake_aws_log_source" "cloudtrail" {
  source {
    accounts = [data.aws_caller_identity.current.account_id]
    regions  = [var.aws_region]
    source_name = "CLOUD_TRAIL_MGMT"
    source_version = "2.0"
  }

  depends_on = [aws_securitylake_data_lake.main]
}

# Custom Log Source for service-logs
resource "aws_securitylake_custom_log_source" "service_logs" {
  source_name    = "service-logs"
  source_version = "1.0"

  configuration {
    crawler_configuration {
      role_arn = aws_iam_role.glue_crawler.arn
    }
    provider_identity {
      external_id = "${var.basename}-service-logs"
      principal   = data.aws_caller_identity.current.account_id
    }
  }

  event_classes = ["ACCESS_ACTIVITY"]

  depends_on = [aws_securitylake_data_lake.main]

  tags = merge(local.common_tags, {
    Name = "${var.basename}-service-logs"
    Type = "custom-log-source"
  })
}

# Data source to get current AWS account ID
data "aws_caller_identity" "current" {}

###########################################
# Glue Crawler for Custom Log Source
###########################################

# IAM Role for Glue Crawler
resource "aws_iam_role" "glue_crawler" {
  name = "${var.basename}-glue-crawler-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "glue.amazonaws.com"
        }
      }
    ]
  })

  tags = merge(local.common_tags, {
    Name = "${var.basename}-glue-crawler-role"
  })
}

# Attach AWSGlueServiceRole policy to Glue Crawler role
resource "aws_iam_role_policy_attachment" "glue_crawler_service" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSGlueServiceRole"
  role       = aws_iam_role.glue_crawler.name
}

# Policy for Glue Crawler to access Security Lake S3 bucket
resource "aws_iam_policy" "glue_crawler_s3" {
  name        = "${var.basename}-glue-crawler-s3-policy"
  description = "Policy for Glue Crawler to access Security Lake S3 bucket"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          "arn:aws:s3:::aws-security-data-lake-${var.aws_region}-${data.aws_caller_identity.current.account_id}",
          "arn:aws:s3:::aws-security-data-lake-${var.aws_region}-${data.aws_caller_identity.current.account_id}/*"
        ]
      }
    ]
  })

  tags = merge(local.common_tags, {
    Name = "${var.basename}-glue-crawler-s3-policy"
  })
}

resource "aws_iam_role_policy_attachment" "glue_crawler_s3" {
  policy_arn = aws_iam_policy.glue_crawler_s3.arn
  role       = aws_iam_role.glue_crawler.name
}

# Glue Crawler - runs daily
resource "aws_glue_crawler" "service_logs" {
  database_name = aws_glue_catalog_database.security_lake.name
  name          = "${var.basename}-service-logs-crawler"
  role          = aws_iam_role.glue_crawler.arn

  s3_target {
    path = "s3://aws-security-data-lake-${var.aws_region}-${data.aws_caller_identity.current.account_id}/ext/service-logs/"
  }

  schedule = "cron(0 2 * * ? *)" # Daily at 2:00 AM UTC

  depends_on = [aws_securitylake_custom_log_source.service_logs]

  tags = merge(local.common_tags, {
    Name = "${var.basename}-service-logs-crawler"
  })
}

# Glue Catalog Database for Security Lake
resource "aws_glue_catalog_database" "security_lake" {
  name        = "${replace(var.basename, "-", "_")}_security_lake"
  description = "Security Lake database for custom log sources"

  tags = merge(local.common_tags, {
    Name = "${var.basename}-security-lake-database"
  })
} 