package goglog

import (
	"bytes"
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	Logger.setDebugOutput(stdoutBuffer, stderrBuffer)
	Logger.setDebugFormatter(&logrus.TextFormatter{DisableTimestamp: true})

	Logger.Debug("Debug")
	Logger.Info("Info")
	Logger.Print("Print")
	Logger.Warn("Warn")
	Logger.Warning("Warning")
	Logger.Error("Error")

	require.EqualValues(`level=info msg=Info
level=info msg=Print
level=warning msg=Warn
level=warning msg=Warning
`, stdoutBuffer.String())
	require.EqualValues(`level=error msg=Error
`, stderrBuffer.String())
}
