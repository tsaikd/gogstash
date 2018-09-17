package cmd

import (
	"context"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"

	// module loader
	_ "github.com/tsaikd/gogstash/modloader"
)

func gogstash(ctx context.Context, confpath string, debug bool) (err error) {
	if debug {
		goglog.Logger.SetLevel(logrus.DebugLevel)
	}

	if runtime.GOMAXPROCS(0) == 1 && runtime.NumCPU() > 1 {
		goglog.Logger.Warnf("set GOMAXPROCS = %d to get better performance", runtime.NumCPU())
	}

	conf, err := config.LoadFromFile(confpath)
	if err != nil {
		return
	}

	if err = conf.Start(ctx); err != nil {
		return
	}

	goglog.Logger.Info("gogstash started...")

	// Check whether any goroutines failed.
	if err = conf.Wait(); err != nil {
		return
	}

	return
}
