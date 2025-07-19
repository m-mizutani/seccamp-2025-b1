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