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
	"github.com/tsaikd/gogstash/config/logevent"

	_ "github.com/tsaikd/gogstash/modloader"
)

var (
	FlagConfig = flagutil.AddStringFlag(cli.StringFlag{
		Name:   "config",
		EnvVar: "CONFIG",
		Value:  "config.json",
		Usage:  "Path to configuration file",
	})
)

func Main() {
	app := cli.NewApp()
	app.Name = "gogstash"
	app.Usage = "Logstash like, written in golang"
	app.Version = version.String()
	app.Action = actionWrapper(MainAction)
	app.Flags = flagutil.AllFlags()
	app.Commands = cmdutil.AllCommands()

	app.Run(os.Args)
}

func MainAction(c *cli.Context) (err error) {
	if runtime.GOMAXPROCS(0) == 1 && runtime.NumCPU() > 1 {
		logger.Warnf("set GOMAXPROCS = %d to get better performance", runtime.NumCPU())
	}

	confpath := c.String(FlagConfig.Name)
	conf, err := config.LoadFromFile(confpath)
	if err != nil {
		return errutil.New("load config failed, "+confpath, err)
	}

	conf.Map(logger)

	evchan := make(chan logevent.LogEvent, 100)
	conf.Map(evchan)

	if _, err = conf.Invoke(conf.RunInputs); err != nil {
		return
	}

	if _, err = conf.Invoke(conf.RunOutputs); err != nil {
		return
	}

	for {
		// infinite sleep
		time.Sleep(1 * time.Hour)
	}
}
