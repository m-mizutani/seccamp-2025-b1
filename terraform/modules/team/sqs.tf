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