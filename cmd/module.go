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
		Default: "",
		Usage:   "Path to configuration file, default search path: config.json, config.yml",
		EnvVar:  "CONFIG",
	}
	flagFollower = &cobrather.BoolFlag{
		Name:    "follower",
		Default: false,
		Usage:   "golang follower mode",
		EnvVar:  "FOLLOWER",
	}
	flagDebug = &cobrather.BoolFlag{
		Name:    "debug",
		Default: false,
		Usage:   "Enable debug logging",
		EnvVar:  "DEBUG",
	}
	flagPProf = &cobrather.StringFlag{
		Name:    "pprof",
		Default: "",
		Usage:   "Enable golang pprof for listening address, ex: localhost:6060",
		EnvVar:  "PPROF",
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
		flagFollower,
		flagDebug,
		flagPProf,
	},
	RunE: func(ctx context.Context, cmd *cobra.Command, args []string) error {
		return gogstash(ctx, flagConfig.String(), flagFollower.Bool(), flagDebug.Bool(), flagPProf.String())
	},
}
