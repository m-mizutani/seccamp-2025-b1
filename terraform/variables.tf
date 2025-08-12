variable "basename" {
  description = "Base name for all resources"
  type        = string
  default     = "seccamp2025-b1"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ap-northeast-1"
}

# Team variable removed - using shared resources only
