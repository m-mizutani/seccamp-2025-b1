###########################################
# Athena Workgroup
###########################################

resource "aws_athena_workgroup" "main" {
  name = "${var.basename}-workgroup"

  configuration {
    enforce_workgroup_configuration    = true
    publish_cloudwatch_metrics_enabled = true

    result_configuration {
      output_location = "s3://${aws_s3_bucket.athena_results.bucket}/results/"

      encryption_configuration {
        encryption_option = "SSE_S3"
      }
    }
  }

  tags = merge(local.common_tags, {
    Name = "${var.basename}-workgroup"
  })
}

###########################################
# Athena Database - REMOVED
# Using Security Lake's auto-created database instead
###########################################

# resource "aws_glue_catalog_database" "main" {
#   name = replace("${var.basename}-db", "-", "_")
# 
#   description = "Database for ${var.basename} OCSF logs"
# }

###########################################
# Athena Named Queries - REMOVED
# These were referencing the old database
###########################################

# resource "aws_athena_named_query" "recent_failures" {
#   name      = "Recent Login Failures"
#   workgroup = aws_athena_workgroup.main.id
#   database  = aws_glue_catalog_database.main.name
#   description = "Find recent login failures in the last 24 hours"
#   
#   query = <<-EOT
# SELECT 
#   time,
#   actor.user.email_addr as user_email,
#   src_endpoint.ip as source_ip,
#   src_endpoint.location.country as country,
#   api.operation
# FROM ocsf_logs
# WHERE api.operation IN ('login_failure', 'suspicious_login')
#   AND time >= CURRENT_TIMESTAMP - INTERVAL '24' HOUR
# ORDER BY time DESC
# LIMIT 100
# EOT
# }
# 
# resource "aws_athena_named_query" "admin_activities" {
#   name      = "Admin Activities"
#   workgroup = aws_athena_workgroup.main.id
#   database  = aws_glue_catalog_database.main.name
#   description = "Monitor admin activities with high severity"
#   
#   query = <<-EOT
# SELECT 
#   time,
#   actor.user.email_addr as admin_email,
#   api.operation,
#   severity_id,
#   api.response.message
# FROM ocsf_logs
# WHERE actor.user.type_id = 2
#   AND severity_id >= 3
#   AND time >= CURRENT_TIMESTAMP - INTERVAL '7' DAY
# ORDER BY time DESC
# LIMIT 100
# EOT
# }
# 
# resource "aws_athena_named_query" "large_downloads" {
#   name      = "Large Download Detection"
#   workgroup = aws_athena_workgroup.main.id
#   database  = aws_glue_catalog_database.main.name
#   description = "Detect users downloading large amounts of data"
#   
#   query = <<-EOT
# WITH hourly_downloads AS (
#   SELECT 
#     actor.user.email_addr as user_email,
#     DATE_TRUNC('hour', time) as hour,
#     COUNT(*) as download_count
#   FROM ocsf_logs
#   WHERE api.operation IN ('download', 'export')
#     AND time >= CURRENT_TIMESTAMP - INTERVAL '24' HOUR
#   GROUP BY actor.user.email_addr, DATE_TRUNC('hour', time)
# )
# SELECT 
#   user_email,
#   hour,
#   download_count
# FROM hourly_downloads
# WHERE download_count > 50
# ORDER BY hour DESC, download_count DESC
# EOT
# }