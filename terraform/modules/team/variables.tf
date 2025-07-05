variable "team_name" {
  description = "Name of the team"
  type        = string
  validation {
    condition = contains([
      "blue", "green", "red", "yellow", "purple", 
      "orange", "pink", "brown", "gray", "black", "white"
    ], var.team_name)
    error_message = "Team name must be one of: blue, green, red, yellow, purple, orange, pink, brown, gray, black, white."
  }
}

variable "basename" {
  description = "Base name for resources"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "common_tags" {
  description = "Common tags to apply to all resources"
  type        = map(string)
  default     = {}
}

variable "security_lake_bucket" {
  description = "Security Lake S3 bucket name"
  type        = string
}

variable "security_lake_database" {
  description = "Security Lake Glue database name"
  type        = string
}

variable "alerts_sns_topic_arn" {
  description = "ARN of the alerts SNS topic"
  type        = string
}

variable "athena_results_bucket" {
  description = "Athena results S3 bucket name"
  type        = string
}

variable "converter_zip_path" {
  description = "Path to the converter Lambda zip file"
  type        = string
}

variable "converter_zip_hash" {
  description = "Hash of the converter Lambda zip file"
  type        = string
} 