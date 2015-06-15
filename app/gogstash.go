package gogstash

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/tsaikd/KDGoLib/flagutil"
	"github.com/tsaikd/KDGoLib/version"
	"github.com/tsaikd/gogstash/config"
)

var (
	FlagDebug = flagutil.AddBoolFlag(cli.BoolFlag{
		Name:   "debug",
		EnvVar: "DEBUG",
		Usage:  "Show DEBUG messages",
	})
	FlagConfig = flagutil.AddStringFlag(cli.StringFlag{
		Name:   "config",
		EnvVar: "CONFIG",
		Value:  "config.json",
		Usage:  "Path to configuration file",
	})
	FlagProfile = flagutil.AddStringFlag(cli.StringFlag{
		Name:   "profile",
		EnvVar: "PROFILE",
		Value:  "",
		Usage:  "Listen http profiling interface, e.g. localhost:6060",
	})
)

func init() {
	version.VERSION = "0.1.0"
}

func Main() {
	app := cli.NewApp()
	app.Name = "gogstash"
	app.Usage = "Logstash like, written in golang"
	app.Version = version.String()
	app.Action = actionWrapper(MainAction)
	app.Flags = flagutil.AllFlags()

	app.Run(os.Args)
}

func MainAction(c *cli.Context) (err error) {
	var (
		conf      config.Config
		eventChan = make(chan config.LogEvent, 100)
	)

	if c.Bool(FlagDebug.Name) {
		log.SetLevel(log.DebugLevel)
	}

	if runtime.GOMAXPROCS(0) == 1 && runtime.NumCPU() > 1 {
		log.Warnf("set GOMAXPROCS = %d to get better performance", runtime.NumCPU())
	}

	confpath := c.String(FlagConfig.Name)
	if conf, err = config.LoadConfig(confpath); err != nil {
		log.Errorf("Load config failed: %q", confpath)
		return
	}

	profile := c.String(FlagProfile.Name)
	if profile != "" {
		go func() {
			log.Infof("Profile listen http: %q", profile)
			if err := http.ListenAndServe(profile, nil); err != nil {
				log.Errorf("Profile listen http failed: %q\n%v", profile, err)
			}
		}()
	}

	for _, input := range conf.Input() {
		log.Debugf("Init input %q", input.Type())
		if err = input.Event(eventChan); err != nil {
			log.Errorf("input failed: %v", err)
		}
	}

	go func() {
		outputs := conf.Output()
		for {
			select {
			case event := <-eventChan:
				for _, output := range outputs {
					if err = output.Event(event); err != nil {
						log.Errorf("output failed: %v", err)
					}
				}
			}
		}
	}()

	for {
		// infinite sleep
		time.Sleep(1 * time.Hour)
	}
}

func actionWrapper(action func(context *cli.Context) error) func(context *cli.Context) {
	return func(context *cli.Context) {
		if err := action(context); err != nil {
			log.Fatal(err)
		}
	}
}
