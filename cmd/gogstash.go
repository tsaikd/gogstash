package cmd

import (
	"context"
	"net/http"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/tsaikd/KDGoLib/futil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"

	// load pprof module
	_ "net/http/pprof"

	// module loader
	_ "github.com/tsaikd/gogstash/modloader"
)

func gogstash(
	ctx context.Context,
	confpath string,
	debug bool,
	pprofAddress string,
) (err error) {
	if debug {
		goglog.Logger.SetLevel(logrus.DebugLevel)
	}

	if runtime.GOMAXPROCS(0) == 1 && runtime.NumCPU() > 1 {
		goglog.Logger.Warnf("set GOMAXPROCS = %d to get better performance", runtime.NumCPU())
	}

	if confpath == "" {
		confpath = searchConfigPath()
	}

	conf, err := config.LoadFromFile(confpath)
	if err != nil {
		return
	}

	if err = conf.Start(ctx); err != nil {
		return
	}

	if pprofAddress != "" {
		go func() {
			if err := http.ListenAndServe(pprofAddress, nil); err != nil {
				goglog.Logger.Error(err)
			}
		}()
	}

	goglog.Logger.Info("gogstash started...")

	// Check whether any goroutines failed.
	if err = conf.Wait(); err != nil {
		return
	}

	return
}

func searchConfigPath() string {
	for _, path := range []string{"config.json", "config.yaml", "config.yml"} {
		if futil.IsExist(path) {
			return path
		}
	}
	return "config.json"
}
