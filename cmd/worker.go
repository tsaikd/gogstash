package cmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/tsaikd/gogstash/config/goglog"
)

func waitSignals(ctx context.Context) error {
	osSignalChan := make(chan os.Signal, 1)
	signal.Notify(osSignalChan, os.Interrupt, os.Kill)

	select {
	case <-ctx.Done():
	case sig := <-osSignalChan:
		goglog.Logger.Info(sig)
	}
	return nil
}
