package main

import (
	"context"
	"fmt"
	"os"

	"github.com/m-mizutani/seccamp-2025-b1/tools/loggen/cmd"
	"github.com/urfave/cli/v3"
)

const version = "v1.0.0"

func main() {
	app := &cli.Command{
		Name:    "loggen",
		Version: version,
		Usage:   "Log Generator for Security Camp 2025 B1",
		Commands: []*cli.Command{
			cmd.GenerateCommand(),
			cmd.ValidateCommand(),
			cmd.PreviewCommand(),
			cmd.StatsCommand(),
			cmd.CompareCommand(),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
