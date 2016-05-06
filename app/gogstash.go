package gogstash

import (
	"os"
	"runtime"
	"time"

	"github.com/codegangsta/cli"
	"github.com/tsaikd/KDGoLib/cliutil/cmdutil"
	"github.com/tsaikd/KDGoLib/cliutil/flagutil"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/KDGoLib/version"
	"github.com/tsaikd/gogstash/config"

	// module loader
	_ "github.com/tsaikd/gogstash/modloader"
)

var (
	flagConfig = flagutil.AddStringFlag(cli.StringFlag{
		Name:   "config",
		EnvVar: "CONFIG",
		Value:  "config.json",
		Usage:  "Path to configuration file",
	})
)

// Main gogstash main entry function
func Main() {
	app := cli.NewApp()
	app.Name = "gogstash"
	app.Usage = "Logstash like, written in golang"
	app.Version = version.String()
	app.Action = actionWrapper(mainAction)
	app.Flags = flagutil.AllFlags()
	app.Commands = cmdutil.AllCommands()

	app.Run(os.Args)
}

func mainAction(c *cli.Context) (err error) {
	if runtime.GOMAXPROCS(0) == 1 && runtime.NumCPU() > 1 {
		logger.Warnf("set GOMAXPROCS = %d to get better performance", runtime.NumCPU())
	}

	confpath := c.String(flagConfig.Name)
	conf, err := config.LoadFromFile(confpath)
	if err != nil {
		return errutil.New("load config failed, "+confpath, err)
	}

	if err = conf.RunInputs(); err != nil {
		return
	}

	if err = conf.RunFilters(); err != nil {
		return
	}

	if err = conf.RunOutputs(); err != nil {
		return
	}

	for {
		// all event run in routine, go into infinite sleep
		time.Sleep(1 * time.Hour)
	}
}
