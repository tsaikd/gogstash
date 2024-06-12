package main

import (
	"log"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/tsaikd/gogstash/cmd"
)

func main() {
	dsn := os.Getenv("GS_SENTRY_DSN")
	if dsn != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              dsn,
			TracesSampleRate: 1.0,
		})
		if err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}
	}
	cmd.Module.MustMainRun()
}
