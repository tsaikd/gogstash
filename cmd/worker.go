package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/tsaikd/gogstash/config/goglog"
)

func waitSignals(ctx context.Context) error {
	osSignalChan := make(chan os.Signal, 1)
	signal.Notify(osSignalChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case sig := <-osSignalChan:
		goglog.Logger.Info(sig)
	}
	return nil
}
