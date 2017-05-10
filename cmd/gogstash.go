package cmd

import (
	"context"
	"runtime"

	"github.com/Sirupsen/logrus"
	"github.com/tsaikd/gogstash/config"

	// module loader
	_ "github.com/tsaikd/gogstash/modloader"
)

func gogstash(ctx context.Context, confpath string, debug bool) (err error) {
	logger := config.Logger

	if debug {
		logger.Level = logrus.DebugLevel
	}

	if runtime.GOMAXPROCS(0) == 1 && runtime.NumCPU() > 1 {
		logger.Warnf("set GOMAXPROCS = %d to get better performance", runtime.NumCPU())
	}

	conf, err := config.LoadFromFile(confpath)
	if err != nil {
		return
	}

	if err = conf.Start(ctx); err != nil {
		return
	}

	logger.Info("gogstash started...")

	// Check whether any goroutines failed.
	if err = conf.Wait(); err != nil {
		return
	}

	return
}
