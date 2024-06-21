package goglog

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
	"github.com/tsaikd/KDGoLib/logrusutil"
	"github.com/tsaikd/KDGoLib/runtimecaller"
)

// Logger app logger
var Logger = newLogger()

const timestampFormat = "2006/01/02 15:04:05"

var logrusFormatter = &logrusutil.ConsoleLogFormatter{
	TimestampFormat:      timestampFormat,
	CallerOffset:         5,
	RuntimeCallerFilters: []runtimecaller.Filter{filterGoglogRuntimeCaller},
}

func filterGoglogRuntimeCaller(callinfo runtimecaller.CallInfo) (valid bool, stop bool) {
	return !strings.Contains(callinfo.PackageName(), "github.com/tsaikd/gogstash/config/goglog"), false
}

// LoggerType struct to hold different Sentry hubs for various log levels
type LoggerType struct {
	stdout         *logrus.Logger
	stderr         *logrus.Logger
	WarnSentryHub  *sentry.Hub
	ErrorSentryHub *sentry.Hub
	FatalSentryHub *sentry.Hub
}

// newLogger initializes and configures the Sentry hubs for logging
func newLogger() *LoggerType {
	warnHub := configureHub(sentry.LevelWarning)
	errorHub := configureHub(sentry.LevelError)
	fatalHub := configureHub(sentry.LevelFatal)

	return &LoggerType{
		stdout: &logrus.Logger{
			Out:       os.Stdout,
			Formatter: logrusFormatter,
			Hooks:     make(logrus.LevelHooks),
			Level:     logrus.InfoLevel,
		},
		stderr: &logrus.Logger{
			Out:       os.Stderr,
			Formatter: logrusFormatter,
			Hooks:     make(logrus.LevelHooks),
			Level:     logrus.InfoLevel,
		},
		WarnSentryHub:  warnHub,
		ErrorSentryHub: errorHub,
		FatalSentryHub: fatalHub,
	}
}

// configureHub clones the current Sentry hub and sets the specified level
func configureHub(level sentry.Level) *sentry.Hub {
	hub := sentry.CurrentHub().Clone()
	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(level)
	})
	return hub
}

// LogWarning logs a warning message to Sentry
func (l *LoggerType) LogWarning(msg string) {
	l.WarnSentryHub.WithScope(func(scope *sentry.Scope) {
		sentry.CaptureMessage(msg)
	})
}

// LogError logs an error message to Sentry
func (l *LoggerType) LogError(msg string) {
	l.ErrorSentryHub.WithScope(func(scope *sentry.Scope) {
		sentry.CaptureMessage(msg)
	})
}

// LogFatal logs a fatal error message to Sentry and exits the program
func (l *LoggerType) LogFatal(msg string) {
	l.FatalSentryHub.WithScope(func(scope *sentry.Scope) {
		sentry.CaptureMessage(msg)
	})
	logrus.Fatal(msg)
}

// WithField wrap logrus function
func (t *LoggerType) WithField(key string, value any) *logrus.Entry {
	return t.stdout.WithField(key, value)
}

// WithFields wrap logrus function
func (t *LoggerType) WithFields(fields logrus.Fields) *logrus.Entry {
	return t.stdout.WithFields(fields)
}

// WithError wrap logrus function
func (t *LoggerType) WithError(err error) *logrus.Entry {
	return t.stdout.WithError(err)
}

// Debugf wrap logrus function
func (t *LoggerType) Debugf(format string, args ...any) {
	t.stdout.Debugf(format, args...)
}

// Infof wrap logrus function
func (t *LoggerType) Infof(format string, args ...any) {
	t.stdout.Infof(format, args...)
}

// Printf wrap logrus function
func (t *LoggerType) Printf(format string, args ...any) {
	t.stdout.Printf(format, args...)
}

// Warnf wrap logrus function
func (t *LoggerType) Warnf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	t.WarnSentryHub.CaptureMessage(msg)
	t.WarnSentryHub.Flush(time.Second * 3)
	t.stdout.Warnf(format, args...)
}

// Warningf wrap logrus function
func (t *LoggerType) Warningf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	t.WarnSentryHub.CaptureMessage(msg)
	t.WarnSentryHub.Flush(time.Second * 3)
	t.stdout.Warningf(format, args...)
}

// Errorf wrap logrus function
func (t *LoggerType) Errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	t.ErrorSentryHub.CaptureMessage(msg)
	t.ErrorSentryHub.Flush(time.Second * 3)
	t.stderr.Errorf(format, args...)
}

// Fatalf wrap logrus function
func (t *LoggerType) Fatalf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	t.FatalSentryHub.CaptureMessage(msg)
	t.FatalSentryHub.Flush(time.Second * 3)
	t.stderr.Fatalf(format, args...)
}

// Panicf wrap logrus function
func (t *LoggerType) Panicf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	t.FatalSentryHub.CaptureMessage(msg)
	t.FatalSentryHub.Flush(time.Second * 3)
	t.stderr.Panicf(format, args...)
}

// Debug wrap logrus function
func (t *LoggerType) Debug(args ...any) {
	t.stdout.Debug(args...)
}

// Info wrap logrus function
func (t *LoggerType) Info(args ...any) {
	t.stdout.Info(args...)
}

// Print wrap logrus function
func (t *LoggerType) Print(args ...any) {
	t.stdout.Print(args...)
}

// Warn wrap logrus function
func (t *LoggerType) Warn(args ...any) {
	t.stdout.Warn(args...)
}

// Warning wrap logrus function
func (t *LoggerType) Warning(args ...any) {
	t.stdout.Warning(args...)
}

// Error wrap logrus function
func (t *LoggerType) Error(args ...any) {
	t.stderr.Error(args...)
}

// Fatal wrap logrus function
func (t *LoggerType) Fatal(args ...any) {
	t.stderr.Fatal(args...)
}

// Panic wrap logrus function
func (t *LoggerType) Panic(args ...any) {
	t.stderr.Panic(args...)
}

// Debugln wrap logrus function
func (t *LoggerType) Debugln(args ...any) {
	t.stdout.Debugln(args...)
}

// Infoln wrap logrus function
func (t *LoggerType) Infoln(args ...any) {
	t.stdout.Infoln(args...)
}

// Println wrap logrus function
func (t *LoggerType) Println(args ...any) {
	t.stdout.Println(args...)
}

// Warnln wrap logrus function
func (t *LoggerType) Warnln(args ...any) {
	t.stdout.Warnln(args...)
}

// Warningln wrap logrus function
func (t *LoggerType) Warningln(args ...any) {
	t.stdout.Warningln(args...)
}

// Errorln wrap logrus function
func (t *LoggerType) Errorln(args ...any) {
	t.stderr.Errorln(args...)
}

// Fatalln wrap logrus function
func (t *LoggerType) Fatalln(args ...any) {
	t.stderr.Fatalln(args...)
}

// Panicln wrap logrus function
func (t *LoggerType) Panicln(args ...any) {
	t.stderr.Panicln(args...)
}

// Panicln wrap logrus function
func (t *LoggerType) Trace(err error) (isErr bool) {
	if err != nil {
		t.Error(err)
		return true
	}
	return false
}

// SetLevel set logger level for filtering output
func (t *LoggerType) SetLevel(level logrus.Level) {
	t.stdout.SetLevel(level)
	t.stderr.SetLevel(level)
}

var _ logrus.FieldLogger = &LoggerType{}
