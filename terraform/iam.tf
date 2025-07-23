###########################################
# IAM Roles and Policies
###########################################

# IAM Role for S3 Raw Logs Access
resource "aws_iam_role" "raw_logs_access" {
  name = "${var.basename}-raw-logs-access-role"

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

  tags = merge(local.common_tags, {
    Name = "${var.basename}-raw-logs-access-role"
  })
}

resource "aws_iam_policy" "raw_logs_access" {
  name        = "${var.basename}-raw-logs-access-policy"
  description = "Policy for accessing raw logs S3 bucket"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject"
        ]
        Resource = "${aws_s3_bucket.raw_logs.arn}/*"
      },
      {
        Effect = "Allow"
        Action = [
          "s3:ListBucket"
        ]
        Resource = aws_s3_bucket.raw_logs.arn
      }
    ]
  })

  tags = merge(local.common_tags, {
    Name = "${var.basename}-raw-logs-access-policy"
  })
}

resource "aws_iam_role_policy_attachment" "raw_logs_access" {
  policy_arn = aws_iam_policy.raw_logs_access.arn
  role       = aws_iam_role.raw_logs_access.name
}

###########################################
# IAM Role for Importer Lambda
###########################################

resource "aws_iam_role" "importer_lambda" {
  name = "${var.basename}-importer-lambda-role"

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

  tags = merge(local.common_tags, {
    Name = "${var.basename}-importer-lambda-role"
    Type = "lambda-role"
  })
}

# Basic Lambda execution role
resource "aws_iam_role_policy_attachment" "importer_lambda_basic" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.importer_lambda.name
}

# S3 permissions for importer Lambda
resource "aws_iam_policy" "importer_lambda_s3" {
  name        = "${var.basename}-importer-lambda-s3-policy"
  description = "S3 permissions for importer Lambda"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.raw_logs.arn,
          "${aws_s3_bucket.raw_logs.arn}/*"
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "importer_lambda_s3" {
  policy_arn = aws_iam_policy.importer_lambda_s3.arn
  role       = aws_iam_role.importer_lambda.name
}

###########################################
# IAM Role for Detector Lambda
###########################################

resource "aws_iam_role" "detector_lambda" {
  name = "${var.basename}-detector-lambda-role"

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

  tags = merge(local.common_tags, {
    Name = "${var.basename}-detector-lambda-role"
    Type = "lambda-role"
  })
}

# Basic Lambda execution role
resource "aws_iam_role_policy_attachment" "detector_lambda_basic" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.detector_lambda.name
}

# Athena permissions for detector Lambda
resource "aws_iam_policy" "detector_lambda_athena" {
  name        = "${var.basename}-detector-lambda-athena-policy"
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
          "athena:StopQueryExecution",
          "athena:GetWorkGroup"
        ]
        Resource = [
          "arn:aws:athena:*:*:workgroup/${aws_athena_workgroup.main.name}"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "glue:GetDatabase",
          "glue:GetTable",
          "glue:GetPartitions"
        ]
        Resource = "*"
      }
    ]
  })

  tags = merge(local.common_tags, {
    Name = "${var.basename}-detector-lambda-athena-policy"
  })
}

resource "aws_iam_role_policy_attachment" "detector_lambda_athena" {
  policy_arn = aws_iam_policy.detector_lambda_athena.arn
  role       = aws_iam_role.detector_lambda.name
}

# S3 permissions for detector Lambda (Athena results bucket)
resource "aws_iam_policy" "detector_lambda_s3" {
  name        = "${var.basename}-detector-lambda-s3-policy"
  description = "S3 permissions for detector Lambda to access Athena results"

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
          aws_s3_bucket.athena_results.arn,
          "${aws_s3_bucket.athena_results.arn}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_securitylake_data_lake.main.s3_bucket_arn,
          "${aws_securitylake_data_lake.main.s3_bucket_arn}/*"
        ]
      }
    ]
  })

  tags = merge(local.common_tags, {
    Name = "${var.basename}-detector-lambda-s3-policy"
  })
}

resource "aws_iam_role_policy_attachment" "detector_lambda_s3" {
  policy_arn = aws_iam_policy.detector_lambda_s3.arn
  role       = aws_iam_role.detector_lambda.name
}

# SNS permissions for detector Lambda (alerts)
resource "aws_iam_policy" "detector_lambda_sns" {
  name        = "${var.basename}-detector-lambda-sns-policy"
  description = "SNS permissions for detector Lambda to send alerts"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sns:Publish"
        ]
        Resource = aws_sns_topic.alerts.arn
      }
    ]
  })

  tags = merge(local.common_tags, {
    Name = "${var.basename}-detector-lambda-sns-policy"
  })
}

resource "aws_iam_role_policy_attachment" "detector_lambda_sns" {
  policy_arn = aws_iam_policy.detector_lambda_sns.arn
  role       = aws_iam_role.detector_lambda.name
} 