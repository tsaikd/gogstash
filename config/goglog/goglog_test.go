package goglog

import (
	"bytes"
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/KDGoLib/logrusutil"
)

func (t *LoggerType) setDebugOutput(stdout io.Writer, stderr io.Writer) {
	t.stdout.Out = stdout
	t.stderr.Out = stderr
}

func (t *LoggerType) setDebugFormatter(formatter logrus.Formatter) {
	t.stdout.Formatter = formatter
	t.stderr.Formatter = formatter
}

func TestLogger(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	stdoutBuffer := &bytes.Buffer{}
	stderrBuffer := &bytes.Buffer{}
	logger := newLogger()

	logger.setDebugOutput(stdoutBuffer, stderrBuffer)
	logger.setDebugFormatter(&logrus.TextFormatter{DisableTimestamp: true})

	logger.Debug("Debug")
	logger.Info("Info")
	logger.Print("Print")
	logger.Warn("Warn")
	logger.Warning("Warning")
	logger.Error("Error")

	require.EqualValues(`level=info msg=Info
level=info msg=Print
level=warning msg=Warn
level=warning msg=Warning
`, stdoutBuffer.String())
	require.EqualValues(`level=error msg=Error
`, stderrBuffer.String())
}

func TestLoggerFileLine(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	stdoutBuffer := &bytes.Buffer{}
	stderrBuffer := &bytes.Buffer{}
	logger := newLogger()

	logger.setDebugOutput(stdoutBuffer, stderrBuffer)
	formatter := &logrusutil.ConsoleLogFormatter{
		TimestampFormat: timestampFormat,
		CallerOffset:    5,
	}
	logger.setDebugFormatter(formatter)

	logger.Info("Info")
	require.Contains(stdoutBuffer.String(), "goglog_test.go:70 [info] Info")
}
