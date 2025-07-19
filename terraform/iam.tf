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