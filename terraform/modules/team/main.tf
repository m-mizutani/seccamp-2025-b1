# Team module configuration without Lambda functions
# Lambda functions are managed separately outside of Terraform

# Reference existing IAM role for the team
data "aws_iam_role" "lambda_detector_role" {
  name = "seccamp2025-b1-detector-lambda-role"
}

# Create IAM user for the team
resource "aws_iam_user" "team_user" {
  name = var.team_id

  tags = {
    Team = var.team_id
  }
}

# Create login profile for console access
resource "aws_iam_user_login_profile" "team_user_login" {
  user                    = aws_iam_user.team_user.name
  password_reset_required = true
}

# Attach ReadOnlyAccess policy
resource "aws_iam_user_policy_attachment" "readonly_access" {
  user       = aws_iam_user.team_user.name
  policy_arn = "arn:aws:iam::aws:policy/ReadOnlyAccess"
}

# Create IAM policy document that matches the detector role's permissions
# Since we can't directly copy policies from a role to a user,
# we'll create a policy that grants similar permissions
resource "aws_iam_user_policy" "detector_permissions" {
  name = "${var.team_id}-detector-permissions"
  user = aws_iam_user.team_user.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "athena:*",
          "glue:GetDatabase",
          "glue:GetDatabases",
          "glue:GetTable",
          "glue:GetTables",
          "glue:GetPartition",
          "glue:GetPartitions",
          "glue:BatchGetPartition",
          "s3:GetObject",
          "s3:ListBucket",
          "s3:PutObject",
          "s3:GetBucketLocation",
          "sns:Publish",
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "lambda:InvokeFunction",
          "lambda:GetFunction",
          "lambda:GetFunctionConfiguration"
        ]
        Resource = "arn:aws:lambda:*:*:function:${var.team_id}-detector"
      },
      {
        Effect = "Allow"
        Action = [
          "lakeformation:GetDataAccess",
          "lakeformation:GetResourceLFTags",
          "lakeformation:ListLFTags",
          "lakeformation:GetLFTag",
          "lakeformation:SearchTablesByLFTags",
          "lakeformation:SearchDatabasesByLFTags"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "securitylake:*"
        ]
        Resource = "*"
      }
    ]
  })
}

# Grant Lake Formation permissions to the IAM user
resource "aws_lakeformation_permissions" "database_permissions" {
  principal   = aws_iam_user.team_user.arn
  permissions = ["DESCRIBE"]

  database {
    name = "amazon_security_lake_glue_db_ap_northeast_1"
  }
}

resource "aws_lakeformation_permissions" "table_permissions" {
  principal   = aws_iam_user.team_user.arn
  permissions = ["SELECT", "DESCRIBE"]

  table {
    database_name = "amazon_security_lake_glue_db_ap_northeast_1"
    name          = "amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0"
  }
}

# Create credentials directory
resource "null_resource" "create_credentials_dir" {
  provisioner "local-exec" {
    command = "mkdir -p ${path.module}/../../../credentials"
  }
}

# Output credentials to JSON file
resource "local_file" "credentials" {
  filename = "${path.module}/../../../credentials/${var.team_id}.json"
  content = jsonencode({
    username     = aws_iam_user.team_user.name
    password     = aws_iam_user_login_profile.team_user_login.password
    console_url  = "https://console.aws.amazon.com/"
    account_id   = data.aws_iam_role.lambda_detector_role.arn != null ? split(":", data.aws_iam_role.lambda_detector_role.arn)[4] : ""
    instructions = "Password reset required on first login"
  })

  depends_on = [
    null_resource.create_credentials_dir,
    aws_iam_user_login_profile.team_user_login
  ]
}

