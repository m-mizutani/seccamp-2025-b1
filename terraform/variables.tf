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

variable "enable_external_subscriber" {
  description = "Enable external subscriber for Security Lake"
  type        = bool
  default     = false
}

# Team variable removed - using shared resources only
