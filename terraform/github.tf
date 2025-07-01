###########################################
# GitHub Actions OIDC Identity Provider
###########################################

# Data source to get GitHub OIDC thumbprint
data "tls_certificate" "github" {
  url = "https://token.actions.githubusercontent.com"
}

# OIDC Identity Provider for GitHub Actions
resource "aws_iam_openid_connect_provider" "github" {
  url             = "https://token.actions.githubusercontent.com"
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.github.certificates[0].sha1_fingerprint]

  tags = merge(local.common_tags, {
    Name = "${var.basename}-github-oidc-provider"
    Type = "oidc-provider"
  })
}

# IAM Role for GitHub Actions
resource "aws_iam_role" "github_actions" {
  name = "${var.basename}-github-actions-role"

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
    Name = "${var.basename}-github-actions-role"
    Type = "github-actions"
  })
}

# Attach AdministratorAccess policy to GitHub Actions role
resource "aws_iam_role_policy_attachment" "github_actions_admin" {
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
  role       = aws_iam_role.github_actions.name
} 