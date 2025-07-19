module github.com/m-mizutani/seccamp-2025-b1/terraform/lambda/auditlog

go 1.21

require (
	github.com/aws/aws-lambda-go v1.41.0
	github.com/m-mizutani/seccamp-2025-b1/internal/logcore v0.0.0
)

replace github.com/m-mizutani/seccamp-2025-b1/internal/logcore => ../../../internal/logcore
