###########################################
# Lambda Functions
###########################################

# Build converter Lambda binary
resource "null_resource" "build_converter" {
  triggers = {
    source_hash = data.archive_file.converter_source.output_base64sha256
  }

  provisioner "local-exec" {
    command = <<-EOT
      cd ${path.module}/lambda/converter
      GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap main.go types.go s3_interface.go convert.go
    EOT
    environment = {
      PAGER = ""
    }
  }
}

# Archive source files for trigger detection
data "archive_file" "converter_source" {
  type        = "zip"
  source_dir  = "${path.module}/lambda/converter"
  output_path = "${path.module}/lambda/converter_source.zip"
  excludes    = ["*.zip", "go.sum", "bootstrap", "*_test.go", "testdata"]
}

# Archive file for converter Lambda
data "archive_file" "converter_lambda_zip" {
  type        = "zip"
  source_dir  = "${path.module}/lambda/converter"
  output_path = "${path.module}/lambda/converter.zip"
  excludes    = ["*.zip", "go.sum", "*_test.go", "testdata", "go.mod"]

  depends_on = [null_resource.build_converter]
}

###########################################
# Converter Lambda (Parquet conversion)
###########################################

# IAM Role for Converter Lambda
resource "aws_iam_role" "converter_lambda" {
  name = "lambda-importer-role"

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
    Name = "lambda-importer-role"
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
  name        = "lambda-importer-sqs-policy"
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
  name        = "lambda-importer-s3-policy"
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
        Resource = "${aws_securitylake_data_lake.main.s3_bucket_arn}/*"
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
  architectures    = ["arm64"]
  timeout          = 300
  memory_size      = 512

  environment {
    variables = {
      SECURITY_LAKE_BUCKET = replace(aws_securitylake_data_lake.main.s3_bucket_arn, "arn:aws:s3:::", "")
      AWS_ACCOUNT_ID       = data.aws_caller_identity.current.account_id
      CUSTOM_LOG_SOURCE    = aws_securitylake_custom_log_source.google_workspace.source_name
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
# AuditLog Lambda (Global shared Lambda)
###########################################

# Build auditlog Lambda binary
resource "null_resource" "build_auditlog" {
  triggers = {
    source_hash = data.archive_file.auditlog_source.output_base64sha256
  }

  provisioner "local-exec" {
    command = <<-EOT
      cd ${path.module}/lambda/auditlog
      GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap main.go
    EOT
    environment = {
      PAGER = ""
    }
  }
}

# Archive source files for trigger detection
data "archive_file" "auditlog_source" {
  type        = "zip"
  source_dir  = "${path.module}/lambda/auditlog"
  output_path = "${path.module}/lambda/auditlog_source.zip"
  excludes    = ["*.zip", "go.sum", "bootstrap", "*_test.go", "testdata"]
}

# Archive file for auditlog Lambda
data "archive_file" "auditlog_lambda_zip" {
  type        = "zip"
  source_dir  = "${path.module}/lambda/auditlog"
  output_path = "${path.module}/lambda/auditlog.zip"
  excludes    = ["*.zip", "go.sum", "*_test.go", "testdata", "go.mod"]

  depends_on = [null_resource.build_auditlog]
}

# IAM Role for AuditLog Lambda
resource "aws_iam_role" "auditlog_lambda" {
  name = "${var.basename}-auditlog-lambda-role"

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
    Name = "${var.basename}-auditlog-lambda-role"
    Type = "lambda-role"
  })
}

# Basic Lambda execution role
resource "aws_iam_role_policy_attachment" "auditlog_lambda_basic" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.auditlog_lambda.name
}

# AuditLog Lambda function
resource "aws_lambda_function" "auditlog" {
  filename         = data.archive_file.auditlog_lambda_zip.output_path
  function_name    = "${var.basename}-auditlog"
  role             = aws_iam_role.auditlog_lambda.arn
  handler          = "bootstrap"
  source_code_hash = data.archive_file.auditlog_lambda_zip.output_base64sha256
  runtime          = "provided.al2"
  architectures    = ["arm64"]
  timeout          = 30
  memory_size      = 256

  depends_on = [
    aws_iam_role_policy_attachment.auditlog_lambda_basic,
  ]

  tags = merge(local.common_tags, {
    Name = "${var.basename}-auditlog"
    Type = "lambda-function"
  })
}

# Lambda Function URL for public access
resource "aws_lambda_function_url" "auditlog" {
  function_name      = aws_lambda_function.auditlog.function_name
  authorization_type = "NONE"

  cors {
    allow_credentials = false
    allow_methods     = ["GET"]
    allow_origins     = ["*"]
    allow_headers     = ["*"]
    max_age          = 86400
  }
}

###########################################
# Importer Lambda (Log Importer)
###########################################

# Build importer Lambda binary
resource "null_resource" "build_importer" {
  triggers = {
    source_hash = data.archive_file.importer_source.output_base64sha256
  }

  provisioner "local-exec" {
    command = <<-EOT
      cd ${path.module}/lambda/importer
      GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap .
    EOT
    environment = {
      PAGER = ""
    }
  }
}

# Archive source files for trigger detection
data "archive_file" "importer_source" {
  type        = "zip"
  source_dir  = "${path.module}/lambda/importer"
  output_path = "${path.module}/lambda/importer_source.zip"
  excludes    = ["*.zip", "go.sum", "bootstrap", "*_test.go", "testdata"]
}

# Archive file for importer Lambda
data "archive_file" "importer_lambda_zip" {
  type        = "zip"
  source_dir  = "${path.module}/lambda/importer"
  output_path = "${path.module}/lambda/importer.zip"
  excludes    = ["*.zip", "go.sum", "*_test.go", "testdata", "go.mod"]

  depends_on = [null_resource.build_importer]
}

# Note: IAM resources for importer Lambda are defined in iam.tf

# Importer Lambda function
resource "aws_lambda_function" "importer" {
  filename         = data.archive_file.importer_lambda_zip.output_path
  function_name    = "${var.basename}-importer"
  role             = aws_iam_role.importer_lambda.arn
  handler          = "bootstrap"
  source_code_hash = data.archive_file.importer_lambda_zip.output_base64sha256
  runtime          = "provided.al2"
  architectures    = ["arm64"]
  timeout          = 300
  memory_size      = 512

  environment {
    variables = {
      AUDITLOG_URL   = aws_lambda_function_url.auditlog.function_url
      S3_BUCKET_NAME = aws_s3_bucket.raw_logs.bucket
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.importer_lambda_basic,
    aws_iam_role_policy_attachment.importer_lambda_s3,
  ]

  tags = merge(local.common_tags, {
    Name = "${var.basename}-importer"
    Type = "lambda-function"
  })
}

# EventBridge rule for 5-minute interval execution
resource "aws_cloudwatch_event_rule" "importer_schedule" {
  name                = "${var.basename}-importer-schedule"
  description         = "Trigger importer Lambda every 5 minutes"
  schedule_expression = "rate(5 minutes)"

  tags = merge(local.common_tags, {
    Name = "${var.basename}-importer-schedule"
    Type = "eventbridge-rule"
  })
}

# EventBridge target
resource "aws_cloudwatch_event_target" "importer_target" {
  rule      = aws_cloudwatch_event_rule.importer_schedule.name
  target_id = "${var.basename}-importer-target"
  arn       = aws_lambda_function.importer.arn
}

# Permission for EventBridge to invoke Lambda
resource "aws_lambda_permission" "allow_eventbridge" {
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.importer.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.importer_schedule.arn
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

resource "aws_s3_bucket_versioning" "athena_results" {
  bucket = aws_s3_bucket.athena_results.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "athena_results" {
  bucket = aws_s3_bucket.athena_results.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "athena_results" {
  bucket = aws_s3_bucket.athena_results.id

  rule {
    id     = "expire-old-results"
    status = "Enabled"

    filter {
      prefix = "results/"
    }

    expiration {
      days = 30
    }
  }
}
