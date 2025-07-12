terraform {
  backend "s3" {
    bucket  = "seccamp2025-b1-terraform"
    key     = "main.tfstate"
    region  = "ap-northeast-1"
    encrypt = true
  }
}
