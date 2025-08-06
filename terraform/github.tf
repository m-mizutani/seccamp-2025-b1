###########################################
# GitHub Actions OIDC Identity Provider
###########################################

# Data source to get GitHub OIDC thumbprint
data "tls_certificate" "github" {
  url = "https://token.actions.githubusercontent.com"
}

# OIDC Identity Provider for GitHub Actions
resource "aws_iam_openid_connect_provider" "github" {
  url            = "https://token.actions.githubusercontent.com"
  client_id_list = ["sts.amazonaws.com"]
  thumbprint_list = distinct(concat(
    [data.tls_certificate.github.certificates[0].sha1_fingerprint],
    [for cert in data.tls_certificate.github.certificates : cert.sha1_fingerprint]
  ))

  tags = merge(local.common_tags, {
    Name = "${var.basename}-github-oidc-provider"
    Type = "oidc-provider"
  })
}

# IAM Role for GitHub Actions - Main Repository (Administrator Access)
resource "aws_iam_role" "github_actions_main" {
  name = "GitHubActionsRole-MainRepo"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = aws_iam_openid_connect_provider.github.arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          }
          StringLike = {
            "token.actions.githubusercontent.com:sub" = "repo:m-mizutani/seccamp-2025-b1:*"
          }
        }
      }
    ]
  })

  tags = merge(local.common_tags, {
    Name = "GitHubActionsRole-MainRepo"
    Type = "github-actions"
  })
}

# Attach AdministratorAccess policy to main repo GitHub Actions role
resource "aws_iam_role_policy_attachment" "github_actions_main_admin" {
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
  role       = aws_iam_role.github_actions_main.name
}

# IAM Role for GitHub Actions - B1-Secmon Repository (Lambda Admin Only)
resource "aws_iam_role" "github_actions_b1_secmon" {
  name = "GitHubActionsRole-B1Secmon"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = aws_iam_openid_connect_provider.github.arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          }
          StringLike = {
            "token.actions.githubusercontent.com:sub" = [
              "repo:seccamp2025-b/b1-secmon:*",
              "repo:seccamp2025-b/b1-secmon:ref:refs/heads/*",
              "repo:seccamp2025-b/b1-secmon:ref:refs/pull/*/merge"
            ]
          }
        }
      }
    ]
  })

  tags = merge(local.common_tags, {
    Name = "GitHubActionsRole-B1Secmon"
    Type = "github-actions"
  })
}

# Attach Lambda-related policies to b1-secmon GitHub Actions role
resource "aws_iam_role_policy_attachment" "github_actions_b1_secmon_lambda_full" {
  policy_arn = "arn:aws:iam::aws:policy/AWSLambda_FullAccess"
  role       = aws_iam_role.github_actions_b1_secmon.name
}

# Attach IAM policy for managing Lambda execution roles
resource "aws_iam_role_policy_attachment" "github_actions_b1_secmon_iam" {
  policy_arn = "arn:aws:iam::aws:policy/IAMReadOnlyAccess"
  role       = aws_iam_role.github_actions_b1_secmon.name
}

# Additional policy for S3 access (Lambda deployment packages)
resource "aws_iam_role_policy" "github_actions_b1_secmon_s3" {
  name = "B1SecmonS3Access"
  role = aws_iam_role.github_actions_b1_secmon.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:ListBucket"
        ]
        Resource = [
          "arn:aws:s3:::seccamp2025-b1-*",
          "arn:aws:s3:::seccamp2025-b1-*/*"
        ]
      }
    ]
  })
} 