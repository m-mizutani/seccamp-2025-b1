###########################################
# Local values
###########################################

locals {
  team_name_title = title(var.team_name)
  
  team_tags = merge(var.common_tags, {
    Team = var.team_name
  })
}

###########################################
# S3 Bucket for raw logs
###########################################

resource "aws_s3_bucket" "raw_logs" {
  bucket = "${var.basename}-${var.team_name}-raw-logs"

  tags = merge(local.team_tags, {
    Name = "${var.basename}-${var.team_name}-raw-logs"
    Type = "s3-bucket"
  })
}

resource "aws_s3_bucket_versioning" "raw_logs" {
  bucket = aws_s3_bucket.raw_logs.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "raw_logs" {
  bucket = aws_s3_bucket.raw_logs.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "raw_logs" {
  bucket = aws_s3_bucket.raw_logs.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

###########################################
# SNS Topic for raw logs
###########################################

resource "aws_sns_topic" "raw_logs" {
  name = "${var.basename}-${var.team_name}-raw-logs-sns"

  tags = merge(local.team_tags, {
    Name = "${var.basename}-${var.team_name}-raw-logs-sns"
    Type = "sns-topic"
  })
}

resource "aws_sns_topic_policy" "raw_logs" {
  arn = aws_sns_topic.raw_logs.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "s3.amazonaws.com"
        }
        Action   = "SNS:Publish"
        Resource = aws_sns_topic.raw_logs.arn
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
        }
      }
    ]
  })
}

###########################################
# SQS Queue for raw logs
###########################################

resource "aws_sqs_queue" "raw_logs_dlq" {
  name = "${var.basename}-${var.team_name}-raw-logs-dlq"

  tags = merge(local.team_tags, {
    Name = "${var.basename}-${var.team_name}-raw-logs-dlq"
    Type = "sqs-queue"
  })
}

resource "aws_sqs_queue" "raw_logs" {
  name = "${var.basename}-${var.team_name}-raw-logs-queue"

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.raw_logs_dlq.arn
    maxReceiveCount     = 3
  })

  visibility_timeout_seconds = 300  # Lambda関数のタイムアウトと合わせる

  tags = merge(local.team_tags, {
    Name = "${var.basename}-${var.team_name}-raw-logs-queue"
    Type = "sqs-queue"
  })
}

resource "aws_sqs_queue_policy" "raw_logs" {
  queue_url = aws_sqs_queue.raw_logs.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "sns.amazonaws.com"
        }
        Action   = "sqs:SendMessage"
        Resource = aws_sqs_queue.raw_logs.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = aws_sns_topic.raw_logs.arn
          }
        }
      }
    ]
  })
}

###########################################
# SNS to SQS Subscription
###########################################

resource "aws_sns_topic_subscription" "raw_logs_sqs" {
  topic_arn = aws_sns_topic.raw_logs.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.raw_logs.arn
}

###########################################
# S3 Bucket Notification
###########################################

resource "aws_s3_bucket_notification" "raw_logs" {
  bucket = aws_s3_bucket.raw_logs.id

  topic {
    topic_arn = aws_sns_topic.raw_logs.arn
    events    = ["s3:ObjectCreated:*"]
  }

  depends_on = [aws_sns_topic_policy.raw_logs]
}

###########################################
# IAM Role for Converter Lambda
###########################################

resource "aws_iam_role" "converter_lambda" {
  name = "lambda-${var.team_name}-importer-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = merge(local.team_tags, {
    Name = "lambda-${var.team_name}-importer-role"
    Type = "lambda-role"
  })
}

# SQS read permissions for converter Lambda
resource "aws_iam_policy" "converter_lambda_sqs" {
  name        = "lambda-${var.team_name}-importer-sqs-policy"
  description = "SQS permissions for converter Lambda"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes"
        ]
        Resource = [
          aws_sqs_queue.raw_logs.arn,
          aws_sqs_queue.raw_logs_dlq.arn
        ]
      }
    ]
  })
}

# S3 permissions for converter Lambda
resource "aws_iam_policy" "converter_lambda_s3" {
  name        = "lambda-${var.team_name}-importer-s3-policy"
  description = "S3 permissions for converter Lambda"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:GetObjectVersion"
        ]
        Resource = "${aws_s3_bucket.raw_logs.arn}/*"
      },
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject",
          "s3:PutObjectAcl"
        ]
        Resource = "${var.security_lake_bucket}/*"
      }
    ]
  })
}

# Attach policies to converter Lambda role
resource "aws_iam_role_policy_attachment" "converter_lambda_basic" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.converter_lambda.name
}

resource "aws_iam_role_policy_attachment" "converter_lambda_sqs" {
  policy_arn = aws_iam_policy.converter_lambda_sqs.arn
  role       = aws_iam_role.converter_lambda.name
}

resource "aws_iam_role_policy_attachment" "converter_lambda_s3" {
  policy_arn = aws_iam_policy.converter_lambda_s3.arn
  role       = aws_iam_role.converter_lambda.name
}

###########################################
# IAM Role for Detector Lambda
###########################################

resource "aws_iam_role" "detector_lambda" {
  name = "lambda-${var.team_name}-detector-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = merge(local.team_tags, {
    Name = "lambda-${var.team_name}-detector-role"
    Type = "lambda-role"
  })
}

# Athena permissions for detector Lambda
resource "aws_iam_policy" "detector_lambda_athena" {
  name        = "lambda-${var.team_name}-detector-athena-policy"
  description = "Athena permissions for detector Lambda"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "athena:StartQueryExecution",
          "athena:GetQueryExecution",
          "athena:GetQueryResults",
          "athena:GetWorkGroup"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "glue:GetDatabase",
          "glue:GetTable",
          "glue:GetPartitions"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          var.security_lake_bucket,
          "${var.security_lake_bucket}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject",
          "s3:GetObject",
          "s3:DeleteObject"
        ]
        Resource = [
          "arn:aws:s3:::${var.athena_results_bucket}/*"
        ]
      }
    ]
  })
}

# SNS permissions for detector Lambda
resource "aws_iam_policy" "detector_lambda_sns" {
  name        = "lambda-${var.team_name}-detector-sns-policy"
  description = "SNS permissions for detector Lambda"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sns:Publish"
        ]
        Resource = var.alerts_sns_topic_arn
      }
    ]
  })
}

# Attach policies to detector Lambda role
resource "aws_iam_role_policy_attachment" "detector_lambda_basic" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.detector_lambda.name
}

resource "aws_iam_role_policy_attachment" "detector_lambda_athena" {
  policy_arn = aws_iam_policy.detector_lambda_athena.arn
  role       = aws_iam_role.detector_lambda.name
}

resource "aws_iam_role_policy_attachment" "detector_lambda_sns" {
  policy_arn = aws_iam_policy.detector_lambda_sns.arn
  role       = aws_iam_role.detector_lambda.name
}

###########################################
# Lambda Functions
###########################################

# Converter Lambda function
resource "aws_lambda_function" "converter" {
  filename         = var.converter_zip_path
  function_name    = "${var.basename}-${var.team_name}-converter"
  role             = aws_iam_role.converter_lambda.arn
  handler          = "bootstrap"
  source_code_hash = var.converter_zip_hash
  runtime          = "provided.al2"
  timeout          = 300
  memory_size      = 512
  architectures    = ["arm64"]

  environment {
    variables = {
      SECURITY_LAKE_BUCKET = var.security_lake_bucket
      AWS_ACCOUNT_ID       = data.aws_caller_identity.current.account_id
      TEAM_NAME            = var.team_name
    }
  }

  tags = merge(local.team_tags, {
    Name = "${var.basename}-${var.team_name}-converter"
    Type = "lambda-function"
  })
}



###########################################
# Lambda Event Sources
###########################################

# SQS trigger for converter Lambda
resource "aws_lambda_event_source_mapping" "converter_sqs" {
  event_source_arn = aws_sqs_queue.raw_logs.arn
  function_name    = aws_lambda_function.converter.arn
  batch_size       = 10
}



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

###########################################
# Data Sources
###########################################

data "aws_caller_identity" "current" {} 