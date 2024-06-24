package outputsentry

import (
	"fmt"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
)

type sentryLogger struct {
	InfoSentryHub  *sentry.Hub
	WarnSentryHub  *sentry.Hub
	FatalSentryHub *sentry.Hub
	ErrorSentryHub *sentry.Hub
}

func initLogger() *sentryLogger {
	sentrySyncTransport := sentry.NewHTTPSyncTransport()
	sentrySyncTransport.Timeout = time.Second * 2
	_ = sentry.Init(sentry.ClientOptions{
		Dsn:              "",
		TracesSampleRate: 1.0,
		Transport:        sentrySyncTransport,
	})

	infoHub := sentry.CurrentHub().Clone()
	infoHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelInfo)
	})

	warnHub := sentry.CurrentHub().Clone()
	warnHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelWarning)
	})

	fatalHub := sentry.CurrentHub().Clone()
	fatalHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelFatal)
	})

	errorHub := sentry.CurrentHub().Clone()
	errorHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelError)
	})

	return &sentryLogger{
		InfoSentryHub:  infoHub,
		WarnSentryHub:  warnHub,
		FatalSentryHub: fatalHub,
		ErrorSentryHub: errorHub,
	}
}

func Test_output_sentry_module(t *testing.T) {
	logger := initLogger()

	// defer sentry.Flush(2 * time.Second)

	for i := 0; i < 60; i++ {
		i := i
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			switch {
			case i%4 == 0:
				logger.InfoSentryHub.CaptureMessage(fmt.Sprintf("HubA Info %d", i))
			case i%4 == 1:
				logger.WarnSentryHub.CaptureMessage(fmt.Sprintf("HubA warn %d", i))
			case i%4 == 2:
				logger.ErrorSentryHub.CaptureMessage(fmt.Sprintf("HubA Error %d", i))
			case i%4 == 3:
				logger.FatalSentryHub.CaptureMessage(fmt.Sprintf("HubA Fatal %d", i))
			}
		})
	}
}
