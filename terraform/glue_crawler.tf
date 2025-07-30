###########################################
# Glue Crawler for Security Lake
###########################################

# Import block for existing Security Lake managed crawler
import {
  to = aws_glue_crawler.security_lake
  id = "google-workspace"
}

# Import existing Security Lake managed crawler
resource "aws_glue_crawler" "security_lake" {
  name = "google-workspace"
  role = aws_iam_role.security_lake_crawler.arn

  # Target the Security Lake S3 bucket
  database_name = "amazon_security_lake_glue_db_${replace(var.aws_region, "-", "_")}"
  
  s3_target {
    path = "s3://${replace(aws_securitylake_data_lake.main.s3_bucket_arn, "arn:aws:s3:::", "")}/ext/google-workspace/"
  }

  # Schedule to run once per day at 2 AM JST (17:00 UTC)
  schedule = "cron(0 17 * * ? *)"

  # Crawler configuration
  configuration = jsonencode({
    Version = 1.0
    Grouping = {
      TableGroupingPolicy = "CombineCompatibleSchemas"
    }
  })

  # Update existing schema rather than creating new tables
  schema_change_policy {
    update_behavior = "UPDATE_IN_DATABASE"
    delete_behavior = "LOG"
  }

  # Crawl all data on each run
  recrawl_policy {
    recrawl_behavior = "CRAWL_EVERYTHING"
  }

  tags = merge(local.common_tags, {
    Name = "google-workspace"
    Type = "glue-crawler"
  })
}