package cmd

import (
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
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return gogstash(flagConfig.String())
	},
}
