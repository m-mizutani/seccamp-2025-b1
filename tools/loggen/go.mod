module github.com/m-mizutani/seccamp-2025-b1/tools/loggen

go 1.24.2

replace (
	github.com/m-mizutani/seccamp-2025-b1 => ../..
	github.com/m-mizutani/seccamp-2025-b1/internal/logcore => ../../internal/logcore
)

require (
	github.com/m-mizutani/seccamp-2025-b1/internal/logcore v0.0.0-00010101000000-000000000000
	github.com/urfave/cli/v3 v3.3.8
)
