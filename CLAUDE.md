# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an educational project for Security Camp 2025 B1 course on "Cloud Platform Security Monitoring Introduction". It demonstrates building a security monitoring platform using AWS services, with a progression from manual to automated monitoring.

## Key Commands

### Terraform Infrastructure Management
```bash
cd terraform
terraform init        # Initialize Terraform
terraform validate    # Validate configuration
terraform plan        # Preview changes
terraform apply       # Apply infrastructure changes

# When using zenv wrapper (if available):
zenv terraform plan
zenv terraform apply
```

### AWS CLI
```bash
cd terraform
zenv aws ...
```

### Lambda Function Development

#### Build Commands
```bash
# Converter Lambda (converts logs to OCSF format)
cd terraform/lambda/converter
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap main.go types.go s3_interface.go convert.go

# AuditLog Lambda (generates test logs)
cd terraform/lambda/auditlog
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap main.go

# Importer Lambda (fetches external logs)
cd terraform/lambda/importer
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap .
```

#### Testing
```bash
# Run tests in any Lambda directory
go test -v ./...

# Run specific test files
go test -v main_test.go
go test -v parquet_test.go
```

### Log Generation Tool
```bash
cd tools/loggen
go build -o loggen .
./loggen generate   # Generate test logs
./loggen validate   # Validate log format
```

## Architecture Overview

### Data Flow Pipeline
```
External API (Google Workspace) 
    � Importer Lambda 
    � S3 Raw Logs Bucket 
    � SNS Notification 
    � SQS Queue 
    � Converter Lambda 
    � OCSF Parquet Format 
    � Security Lake S3 Bucket 
    � Glue Crawler 
    � Athena Queries
```

### Key Components

1. **Security Lake**: Central repository for security logs in OCSF format
   - Path format: `ext/{source}/1.0/region={region}/accountId={accountId}/eventDay={YYYYMMDD}/`
   - Custom log source: "google-workspace"

2. **Lambda Functions** (Go-based, ARM64):
   - **Importer**: Fetches logs from external APIs, outputs JSONL format
   - **Converter**: Transforms JSONL to OCSF Parquet format for Security Lake
   - **AuditLog**: Generates simulated logs for testing

3. **Infrastructure**:
   - All resources prefixed with `seccamp2025-b1`
   - S3 backend for Terraform state
   - GitHub Actions OIDC for CI/CD
   - SNS/SQS for event-driven processing

### Important Technical Details

1. **OCSF Schema**: Using Web Resource Activity (Class ID: 6001)
   - Parquet field names must be unique (watch for duplicate "region" fields)
   - Required fields: category_uid, class_uid, type_uid, activity_id, severity_id, time

2. **File Processing**:
   - Input files may be gzip compressed (`.jsonl.gz`)
   - Output must be `.parquet` format (not `.gz.parquet`)
   - Converter handles decompression automatically

3. **Security Lake Integration**:
   - Bucket name includes random ID: `aws-security-data-lake-{region}-{random}`
   - Use Terraform attribute `aws_securitylake_data_lake.main.s3_bucket_arn`
   - Never hardcode bucket names

4. **Common Issues**:
   - SQS messages are wrapped in SNS format - parse accordingly
   - Glue table may have duplicate column names - ensure unique Parquet field names
   - KMS encryption can be complex - use S3_MANAGED_KEY instead

## Development Workflow

1. Make code changes to Lambda functions
2. Terraform automatically rebuilds and deploys on `apply`
3. Upload test files to raw logs bucket to trigger processing
4. Check CloudWatch Logs for Lambda execution logs
5. Run Glue Crawler after new data is added
6. Query data using Athena in the Security Lake database