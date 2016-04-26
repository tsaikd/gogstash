package config

import (
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/KDGoLib/logutil"
)

var (
	// Logger app logger
	Logger = logutil.DefaultLogger
)

func init() {
	formatter := errutil.NewConsoleFormatter("; ")
	errutil.SetDefaultFormatter(formatter)
}
