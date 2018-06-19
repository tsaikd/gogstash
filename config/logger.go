package config

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/KDGoLib/logutil"
)

var (
	// Logger app logger
	Logger = &logrus.Logger{
		Out:       os.Stdout,
		Formatter: logutil.DefaultLogger.Formatter,
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}

	// ErrorLogger app logger to stderr
	ErrorLogger = &logrus.Logger{
		Out:       os.Stderr,
		Formatter: logutil.DefaultLogger.Formatter,
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}
)

func init() {
	formatter := errutil.NewConsoleFormatter("; ")
	errutil.SetDefaultFormatter(formatter)
}
