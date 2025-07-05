variable "basename" {
  description = "Base name for all resources"
  type        = string
  default     = "seccamp2025-b1-poc"
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

variable "teams" {
  description = "List of team names"
  type        = list(string)
  default     = ["blue", "green"]  # Start with one team for testing
}
