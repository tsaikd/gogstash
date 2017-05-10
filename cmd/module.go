package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/tsaikd/KDGoLib/cliutil/cobrather"
)

// command line flags
var (
	flagConfig = &cobrather.StringFlag{
		Name:    "config",
		Default: "config.json",
		Usage:   "Path to configuration file",
		EnvVar:  "CONFIG",
	}
	flagDebug = &cobrather.BoolFlag{
		Name:    "debug",
		Default: false,
		Usage:   "Enable debug logging",
		EnvVar:  "DEBUG",
	}
)

// Module info
var Module = &cobrather.Module{
	Use:   "gogstash",
	Short: "Logstash like, written in golang",
	Commands: []*cobrather.Module{
		cobrather.VersionModule,
	},
	Flags: []cobrather.Flag{
		flagConfig,
		flagDebug,
	},
	RunE: func(ctx context.Context, cmd *cobra.Command, args []string) error {
		return gogstash(ctx, flagConfig.String(), flagDebug.Bool())
	},
}
