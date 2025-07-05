###########################################
# Lambda Functions
###########################################

# Archive file for converter Lambda
data "archive_file" "converter_lambda_zip" {
  type        = "zip"
  source_dir  = "${path.module}/lambda/converter"
  output_path = "${path.module}/lambda/converter.zip"
  excludes    = ["*.zip", "go.sum"]
}

# Archive file for detector Lambda
data "archive_file" "detector_lambda_zip" {
  type        = "zip"
  source_dir  = "${path.module}/lambda/detector"
  output_path = "${path.module}/lambda/detector.zip"
  excludes    = ["*.zip", "go.sum"]
}

###########################################
# Converter Lambda (Parquet conversion)
###########################################

# IAM Role for Converter Lambda
resource "aws_iam_role" "converter_lambda" {
  name = "${var.basename}-converter-lambda-role"

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
    Name = "${var.basename}-converter-lambda-role"
    Type = "lambda-role"
  })
}

# Basic Lambda execution role
resource "aws_iam_role_policy_attachment" "converter_lambda_basic" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.converter_lambda.name
}

# SQS read permissions for converter Lambda
resource "aws_iam_policy" "converter_lambda_sqs" {
  name        = "${var.basename}-converter-lambda-sqs-policy"
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
        Resource = aws_sqs_queue.raw_logs.arn
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "converter_lambda_sqs" {
  policy_arn = aws_iam_policy.converter_lambda_sqs.arn
  role       = aws_iam_role.converter_lambda.name
}

# S3 permissions for converter Lambda
resource "aws_iam_policy" "converter_lambda_s3" {
  name        = "${var.basename}-converter-lambda-s3-policy"
  description = "S3 permissions for converter Lambda"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject"
        ]
        Resource = "${aws_s3_bucket.raw_logs.arn}/*"
      },
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject"
        ]
        Resource = "arn:aws:s3:::aws-security-data-lake-${var.aws_region}-*/*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "converter_lambda_s3" {
  policy_arn = aws_iam_policy.converter_lambda_s3.arn
  role       = aws_iam_role.converter_lambda.name
}

# Converter Lambda function
resource "aws_lambda_function" "converter" {
  filename         = data.archive_file.converter_lambda_zip.output_path
  function_name    = "${var.basename}-converter"
  role             = aws_iam_role.converter_lambda.arn
  handler          = "main"
  source_code_hash = data.archive_file.converter_lambda_zip.output_base64sha256
  runtime          = "provided.al2"
  timeout          = 300
  memory_size      = 512

  environment {
    variables = {
      SECURITY_LAKE_BUCKET = "aws-security-data-lake-${var.aws_region}-${data.aws_caller_identity.current.account_id}"
      AWS_ACCOUNT_ID       = data.aws_caller_identity.current.account_id
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.converter_lambda_basic,
    aws_iam_role_policy_attachment.converter_lambda_sqs,
    aws_iam_role_policy_attachment.converter_lambda_s3,
  ]

  tags = merge(local.common_tags, {
    Name = "${var.basename}-converter"
    Type = "lambda-function"
  })
}

# SQS trigger for converter Lambda
resource "aws_lambda_event_source_mapping" "converter_sqs" {
  event_source_arn = aws_sqs_queue.raw_logs.arn
  function_name    = aws_lambda_function.converter.arn
  batch_size       = 1
}

###########################################
# Detector Lambda (Alert detection)
###########################################

# IAM Role for Detector Lambda
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

# Basic Lambda execution role for detector
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
          "athena:StopQueryExecution"
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
          "arn:aws:s3:::aws-security-data-lake-${var.aws_region}-${data.aws_caller_identity.current.account_id}",
          "arn:aws:s3:::aws-security-data-lake-${var.aws_region}-${data.aws_caller_identity.current.account_id}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject",
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          "arn:aws:s3:::${var.basename}-athena-results",
          "arn:aws:s3:::${var.basename}-athena-results/*"
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
}

resource "aws_iam_role_policy_attachment" "detector_lambda_athena" {
  policy_arn = aws_iam_policy.detector_lambda_athena.arn
  role       = aws_iam_role.detector_lambda.name
}

# SNS permissions for detector Lambda
resource "aws_iam_policy" "detector_lambda_sns" {
  name        = "${var.basename}-detector-lambda-sns-policy"
  description = "SNS permissions for detector Lambda"

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
}

resource "aws_iam_role_policy_attachment" "detector_lambda_sns" {
  policy_arn = aws_iam_policy.detector_lambda_sns.arn
  role       = aws_iam_role.detector_lambda.name
}

# S3 bucket for Athena query results
resource "aws_s3_bucket" "athena_results" {
  bucket = "${var.basename}-athena-results"

  tags = merge(local.common_tags, {
    Name = "${var.basename}-athena-results"
    Type = "athena-results"
  })
}

resource "aws_s3_bucket_public_access_block" "athena_results" {
  bucket = aws_s3_bucket.athena_results.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Detector Lambda function
resource "aws_lambda_function" "detector" {
  filename         = data.archive_file.detector_lambda_zip.output_path
  function_name    = "${var.basename}-detector"
  role             = aws_iam_role.detector_lambda.arn
  handler          = "main"
  source_code_hash = data.archive_file.detector_lambda_zip.output_base64sha256
  runtime          = "provided.al2"
  timeout          = 900 # 15 minutes for query execution
  memory_size      = 512

  environment {
    variables = {
      ALERTS_SNS_TOPIC_ARN  = aws_sns_topic.alerts.arn
      ATHENA_DATABASE       = "amazon_security_lake_glue_db_${replace(var.aws_region, "-", "_")}"
      ATHENA_RESULTS_BUCKET = aws_s3_bucket.athena_results.id
      AWS_ACCOUNT_ID        = data.aws_caller_identity.current.account_id
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.detector_lambda_basic,
    aws_iam_role_policy_attachment.detector_lambda_athena,
    aws_iam_role_policy_attachment.detector_lambda_sns,
  ]

  tags = merge(local.common_tags, {
    Name = "${var.basename}-detector"
    Type = "lambda-function"
  })
}

# EventBridge rule for periodic execution of detector Lambda
resource "aws_cloudwatch_event_rule" "detector_schedule" {
  name                = "${var.basename}-detector-schedule"
  description         = "Trigger detector Lambda every hour"
  schedule_expression = "rate(1 hour)"

  tags = merge(local.common_tags, {
    Name = "${var.basename}-detector-schedule"
  })
}

# EventBridge target for detector Lambda
resource "aws_cloudwatch_event_target" "detector_lambda" {
  rule      = aws_cloudwatch_event_rule.detector_schedule.name
  target_id = "DetectorLambdaTarget"
  arn       = aws_lambda_function.detector.arn
}

# Permission for EventBridge to invoke detector Lambda
resource "aws_lambda_permission" "allow_eventbridge" {
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.detector.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.detector_schedule.arn
} 