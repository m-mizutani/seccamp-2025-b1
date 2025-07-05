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
# SNS to SQS Subscription
###########################################

resource "aws_sns_topic_subscription" "raw_logs_sqs" {
  topic_arn = aws_sns_topic.raw_logs.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.raw_logs.arn
} 