###########################################
# SNS/SQS Notification System
###########################################

# SNS Topic for Raw Logs
resource "aws_sns_topic" "raw_logs" {
  name = "${var.basename}-raw-logs-sns"

  # KMS encryption disabled to allow S3 to publish messages
  # kms_master_key_id = "alias/aws/sns"

  tags = merge(local.common_tags, {
    Name = "${var.basename}-raw-logs-sns"
    Type = "notification"
  })
}

# SQS Queue for Raw Logs
resource "aws_sqs_queue" "raw_logs_dlq" {
  name = "${var.basename}-raw-logs-dlq"

  tags = merge(local.common_tags, {
    Name = "${var.basename}-raw-logs-dlq"
    Type = "dead-letter-queue"
  })
}

resource "aws_sqs_queue" "raw_logs" {
  name                       = "${var.basename}-raw-logs-queue"
  visibility_timeout_seconds = 300

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.raw_logs_dlq.arn
    maxReceiveCount     = 3
  })

  tags = merge(local.common_tags, {
    Name = "${var.basename}-raw-logs-queue"
    Type = "notification-queue"
  })
}

# SQS Queue Policy to allow SNS to send messages
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

# SNS to SQS Subscription
resource "aws_sns_topic_subscription" "raw_logs_sqs" {
  topic_arn = aws_sns_topic.raw_logs.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.raw_logs.arn
}

# SNS Topic Policy to allow S3 to publish messages
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
        Action   = "sns:Publish"
        Resource = aws_sns_topic.raw_logs.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = aws_s3_bucket.raw_logs.arn
          }
        }
      }
    ]
  })
}

# SNS Topic for Alerts
resource "aws_sns_topic" "alerts" {
  name = "${var.basename}-alerts-sns"

  kms_master_key_id = "alias/aws/sns"

  tags = merge(local.common_tags, {
    Name = "${var.basename}-alerts-sns"
    Type = "alerts"
  })
} 