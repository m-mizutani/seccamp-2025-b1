###########################################
# Security Lake Custom Log Source
###########################################

resource "aws_securitylake_custom_log_source" "team_logs" {
  source_name    = var.team_name
  source_version = "1.0"

  event_classes = ["AUTHENTICATION", "AUTHORIZATION", "NETWORK_ACTIVITY"]

  configuration {
    crawler_configuration {
      role_arn = aws_iam_role.security_lake_crawler.arn
    }
    provider_identity {
      external_id = "${var.team_name}-logs-${random_id.external_id.hex}"
      principal   = data.aws_caller_identity.current.account_id
    }
  }
}

# Random ID for external ID
resource "random_id" "external_id" {
  byte_length = 8
}

# IAM Role for Security Lake Crawler
resource "aws_iam_role" "security_lake_crawler" {
  name = "${var.basename}-${var.team_name}-security-lake-crawler-role"

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

  tags = merge(local.team_tags, {
    Name = "${var.basename}-${var.team_name}-security-lake-crawler-role"
    Type = "iam-role"
  })
}

# Policy for Security Lake Crawler
resource "aws_iam_policy" "security_lake_crawler" {
  name        = "${var.basename}-${var.team_name}-security-lake-crawler-policy"
  description = "Policy for Security Lake Crawler to access S3 and Glue"

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
          aws_s3_bucket.raw_logs.arn,
          "${aws_s3_bucket.raw_logs.arn}/*"
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
          "lakeformation:ListPermissions",
          "lakeformation:GetLFTag",
          "lakeformation:ListLFTags",
          "lakeformation:GetResourceLFTags",
          "lakeformation:ListResourceLFTags"
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