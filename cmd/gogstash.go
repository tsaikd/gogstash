package cmd

import (
	"context"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/getsentry/sentry-go"
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
	workerMode bool,
) (err error) {
	defer func() {
		if err != nil {
			hub := sentry.CurrentHub().Clone()
			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetLevel(sentry.LevelError)
			})
			hub.CaptureException(err)
		}
	}()

	dsn := os.Getenv("GS_SENTRY_DSN")

	if dsn != "" {
		sentrySyncTransport := sentry.NewHTTPSyncTransport()
		sentrySyncTransport.Timeout = time.Second * 3

		err := sentry.Init(sentry.ClientOptions{
			Dsn:              dsn,
			TracesSampleRate: 1.0,
			Transport:        sentrySyncTransport,
		})

		if err != nil {
			return err
		}
	}

	goglog.NewLogger()

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
		return err
	}

	// use worker mode when user need more than one worker
	if conf.Worker > 1 && !workerMode {
		return startWorkers(ctx, conf.Worker)
	}

	if err := conf.Start(ctx); err != nil {
		return err
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
	return conf.Wait()
}

func searchConfigPath() string {
	for _, path := range []string{"config.json", "config.yaml", "config.yml"} {
		if futil.IsExist(path) {
			return path
		}
	}
	return "config.json"
}
