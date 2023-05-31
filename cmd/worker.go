package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/tsaikd/gogstash/config/goglog"
)

func waitSignals(ctx context.Context) os.Signal {
	osSignalChan := make(chan os.Signal, 1)
	signal.Notify(osSignalChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		return nil
	case sig := <-osSignalChan:
		goglog.Logger.Info(sig)
		return sig
	}
}
