###########################################
# S3 Bucket for Raw Logs
###########################################

resource "aws_s3_bucket" "raw_logs" {
  bucket = "${var.basename}-raw-logs"

  tags = merge(local.common_tags, {
    Name = "${var.basename}-raw-logs"
    Type = "raw-logs"
  })
}

resource "aws_s3_bucket_public_access_block" "raw_logs" {
  bucket = aws_s3_bucket.raw_logs.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
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

# S3 Bucket Notification Configuration
resource "aws_s3_bucket_notification" "raw_logs" {
  bucket = aws_s3_bucket.raw_logs.id

  topic {
    topic_arn = aws_sns_topic.raw_logs.arn
    events    = [
      "s3:ObjectCreated:Put",
      "s3:ObjectCreated:Post",
      "s3:ObjectCreated:Copy",
      "s3:ObjectCreated:CompleteMultipartUpload"
    ]
  }

  depends_on = [aws_sns_topic_policy.raw_logs]
} 