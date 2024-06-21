package outputsentry

import (
	"fmt"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/require"
)

type sentryLogger struct {
	InfoSentryHub  func(string, ...any)
	WarnSentryHub  func(string, ...any)
	FatalSentryHub func(string, ...any)
	ErrorSentryHub func(string, ...any)
}

func initLogger() *sentryLogger {
	_ = sentry.Init(sentry.ClientOptions{
		Dsn:              "https://efc659a87d984025d9e3810cd6cdee8d@sr.cdmx.io/17",
		TracesSampleRate: 1.0,
	})

	infoHub := func(format string, args ...any) {
		hub := sentry.CurrentHub().Clone()
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetLevel(sentry.LevelInfo)
		})
		hub.CaptureMessage(fmt.Sprintf(format, args...))
		hub.Flush(time.Second * 3)
	}

	warnHub := func(format string, args ...any) {
		hub := sentry.CurrentHub().Clone()
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetLevel(sentry.LevelWarning)
		})
		hub.CaptureMessage(fmt.Sprintf(format, args...))
		hub.Flush(time.Second * 3)
	}

	fatalHub := func(format string, args ...any) {
		hub := sentry.CurrentHub().Clone()
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetLevel(sentry.LevelFatal)
		})
		hub.CaptureMessage(fmt.Sprintf(format, args...))
		hub.Flush(time.Second * 3)
	}

	errorHub := func(format string, args ...any) {
		hub := sentry.CurrentHub().Clone()
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetLevel(sentry.LevelError)
		})
		hub.CaptureMessage(fmt.Sprintf(format, args...))
		hub.Flush(time.Second * 3)
	}

	return &sentryLogger{
		InfoSentryHub:  infoHub,
		WarnSentryHub:  warnHub,
		FatalSentryHub: fatalHub,
		ErrorSentryHub: errorHub,
	}
}

func Test_output_sentry_module(t *testing.T) {
	logger := initLogger()

	for i := 0; i < 60; i++ {
		i := i
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			switch {
			case i%4 == 0:
				logger.InfoSentryHub(fmt.Sprintf("Hub Info %d", i))
			case i%4 == 1:
				logger.WarnSentryHub(fmt.Sprintf("Hub warn %d", i))
			case i%4 == 2:
				logger.ErrorSentryHub(fmt.Sprintf("Hub Error %d", i))
			case i%4 == 3:
				logger.FatalSentryHub(fmt.Sprintf("Hub Fatal %d", i))
			}
		})
	}
}

func TestSentryRecover(t *testing.T) {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              "", // Use an actual DSN here
		TracesSampleRate: 1.0,
	})
	require.NoError(t, err)

	require.NotPanics(t, func() {
		defer func() {
			if err := recover(); err != nil {
				// Report the panic to Sentry
				sentry.CurrentHub().Recover(err)
				// Flush the buffered events to ensure the panic is sent to Sentry
				sentry.Flush(time.Second * 5)
			}
		}()
		// Code that causes a panic
		panic("panic")
	})
}
