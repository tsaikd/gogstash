package goglog

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
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

func newLogger() *LoggerType {
	logger := &LoggerType{
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
	}
	logger.ConfigSentry(sentry.CurrentHub())
	return logger
}

type LoggerType struct {
	mutex  sync.Mutex
	stdout *logrus.Logger
	stderr *logrus.Logger

	debugHub *sentry.Hub
	infoHub  *sentry.Hub
	warnHub  *sentry.Hub
	errorHub *sentry.Hub
	fatalHub *sentry.Hub
}

func (t *LoggerType) ConfigSentry(hub *sentry.Hub) {
	debugHub := hub.Clone()
	debugHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelDebug)
	})
	infoHub := hub.Clone()
	infoHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelInfo)
	})
	warnHub := hub.Clone()
	warnHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelWarning)
	})
	errorHub := hub.Clone()
	errorHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelError)
	})
	fatalHub := hub.Clone()
	fatalHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelFatal)
	})

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.debugHub = debugHub
	t.infoHub = infoHub
	t.warnHub = warnHub
	t.errorHub = errorHub
	t.fatalHub = fatalHub
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
	t.debugHub.CaptureMessage(fmt.Sprintf(format, args...))
}

// Infof wrap logrus function
func (t *LoggerType) Infof(format string, args ...any) {
	t.stdout.Infof(format, args...)
	t.infoHub.CaptureMessage(fmt.Sprintf(format, args...))
}

// Printf wrap logrus function
func (t *LoggerType) Printf(format string, args ...any) {
	t.stdout.Printf(format, args...)
	t.infoHub.CaptureMessage(fmt.Sprintf(format, args...))
}

// Warnf wrap logrus function
func (t *LoggerType) Warnf(format string, args ...any) {
	t.stdout.Warnf(format, args...)
	t.warnHub.CaptureException(fmt.Errorf(format, args...))
}

// Warningf wrap logrus function
func (t *LoggerType) Warningf(format string, args ...any) {
	t.stdout.Warningf(format, args...)
	t.warnHub.CaptureException(fmt.Errorf(format, args...))
}

// Errorf wrap logrus function
func (t *LoggerType) Errorf(format string, args ...any) {
	t.stderr.Errorf(format, args...)
	t.errorHub.CaptureException(fmt.Errorf(format, args...))
}

// Fatalf wrap logrus function
func (t *LoggerType) Fatalf(format string, args ...any) {
	t.fatalHub.CaptureException(fmt.Errorf(format, args...))
	t.stderr.Fatalf(format, args...)
}

// Panicf wrap logrus function
func (t *LoggerType) Panicf(format string, args ...any) {
	t.fatalHub.CaptureException(fmt.Errorf(format, args...))
	t.stderr.Panicf(format, args...)
}

// Debug wrap logrus function
func (t *LoggerType) Debug(args ...any) {
	t.stdout.Debug(args...)
	t.debugHub.CaptureMessage(fmt.Sprint(args...))
}

// Info wrap logrus function
func (t *LoggerType) Info(args ...any) {
	t.stdout.Info(args...)
	t.infoHub.CaptureMessage(fmt.Sprint(args...))
}

// Print wrap logrus function
func (t *LoggerType) Print(args ...any) {
	t.stdout.Print(args...)
	t.infoHub.CaptureMessage(fmt.Sprint(args...))
}

// Warn wrap logrus function
func (t *LoggerType) Warn(args ...any) {
	t.stdout.Warn(args...)
	t.warnHub.CaptureException(errors.New(fmt.Sprint(args...)))
}

// Warning wrap logrus function
func (t *LoggerType) Warning(args ...any) {
	t.stdout.Warning(args...)
	t.warnHub.CaptureException(errors.New(fmt.Sprint(args...)))
}

// Error wrap logrus function
func (t *LoggerType) Error(args ...any) {
	t.stderr.Error(args...)
	t.errorHub.CaptureException(errors.New(fmt.Sprint(args...)))
}

// Fatal wrap logrus function
func (t *LoggerType) Fatal(args ...any) {
	t.fatalHub.CaptureException(errors.New(fmt.Sprint(args...)))
	t.stderr.Fatal(args...)
}

// Panic wrap logrus function
func (t *LoggerType) Panic(args ...any) {
	t.fatalHub.CaptureException(errors.New(fmt.Sprint(args...)))
	t.stderr.Panic(args...)
}

// Debugln wrap logrus function
func (t *LoggerType) Debugln(args ...any) {
	t.stdout.Debugln(args...)
	t.debugHub.CaptureMessage(t.sprintlnn(args...))
}

// Infoln wrap logrus function
func (t *LoggerType) Infoln(args ...any) {
	t.stdout.Infoln(args...)
	t.infoHub.CaptureMessage(t.sprintlnn(args...))
}

// Println wrap logrus function
func (t *LoggerType) Println(args ...any) {
	t.stdout.Println(args...)
	t.infoHub.CaptureMessage(t.sprintlnn(args...))
}

// Warnln wrap logrus function
func (t *LoggerType) Warnln(args ...any) {
	t.stdout.Warnln(args...)
	t.warnHub.CaptureException(errors.New(t.sprintlnn(args...)))
}

// Warningln wrap logrus function
func (t *LoggerType) Warningln(args ...any) {
	t.stdout.Warningln(args...)
	t.warnHub.CaptureException(errors.New(t.sprintlnn(args...)))
}

// Errorln wrap logrus function
func (t *LoggerType) Errorln(args ...any) {
	t.stderr.Errorln(args...)
	t.errorHub.CaptureException(errors.New(t.sprintlnn(args...)))
}

// Fatalln wrap logrus function
func (t *LoggerType) Fatalln(args ...any) {
	t.fatalHub.CaptureException(errors.New(t.sprintlnn(args...)))
	t.stderr.Fatalln(args...)
}

// Panicln wrap logrus function
func (t *LoggerType) Panicln(args ...any) {
	t.fatalHub.CaptureException(errors.New(t.sprintlnn(args...)))
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

func (t *LoggerType) FlushSentry(timeout time.Duration) {
	sentry.CurrentHub().Flush(timeout)
	t.debugHub.Flush(timeout)
	t.infoHub.Flush(timeout)
	t.warnHub.Flush(timeout)
	t.errorHub.Flush(timeout)
	t.fatalHub.Flush(timeout)
}

// SetLevel set logger level for filtering output
func (t *LoggerType) SetLevel(level logrus.Level) {
	t.stdout.SetLevel(level)
	t.stderr.SetLevel(level)
}

func (t *LoggerType) sprintlnn(args ...any) string {
	msg := fmt.Sprintln(args...)
	return msg[:len(msg)-1]
}

var _ logrus.FieldLogger = &LoggerType{}
