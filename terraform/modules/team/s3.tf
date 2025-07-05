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