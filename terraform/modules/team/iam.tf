###########################################
# IAM Role for Importer Lambda (S3 to SNS)
###########################################

resource "aws_iam_role" "importer_lambda" {
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

# S3 permissions for importer Lambda (raw log bucket access)
resource "aws_iam_policy" "importer_lambda_s3" {
  name        = "lambda-${var.team_name}-importer-s3-policy"
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

# Attach policies to importer Lambda role
resource "aws_iam_role_policy_attachment" "importer_lambda_basic" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.importer_lambda.name
}

resource "aws_iam_role_policy_attachment" "importer_lambda_s3" {
  policy_arn = aws_iam_policy.importer_lambda_s3.arn
  role       = aws_iam_role.importer_lambda.name
}

###########################################
# IAM Role for Converter Lambda (SQS to Security Lake)
###########################################

resource "aws_iam_role" "converter_lambda" {
  name = "lambda-${var.team_name}-converter-role"

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
    Name = "lambda-${var.team_name}-converter-role"
    Type = "lambda-role"
  })
}

# SQS read permissions for converter Lambda
resource "aws_iam_policy" "converter_lambda_sqs" {
  name        = "lambda-${var.team_name}-converter-sqs-policy"
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
  name        = "lambda-${var.team_name}-converter-s3-policy"
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