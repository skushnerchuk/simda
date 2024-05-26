package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSLogger(t *testing.T) {
	fmt.Print("\nSLogger tests:\n\n")
	t.Run("check level", func(t *testing.T) {
		var b bytes.Buffer
		l := NewSLogger(&b, LevelError)

		l.Info("Info message")
		require.Empty(t, b.String())

		l.Warn("Warn message")
		require.Empty(t, b.String())

		l.Debug("Debug message")
		require.Empty(t, b.String())

		l.Error("Error message")

		logRecord := make(map[string]interface{})
		err := json.Unmarshal(b.Bytes(), &logRecord)
		require.NoError(t, err)
		require.Equal(t, "ERROR", logRecord["level"])
		require.Equal(t, "Error message", logRecord["msg"])
	})

	t.Run("check enable/disable", func(t *testing.T) {
		var b bytes.Buffer
		l := NewSLogger(&b, LevelInfo)

		l.Disable()
		l.Info("Info message")
		require.Empty(t, b.String())

		l.Enable()
		l.Info("Info message")
		require.NotEmpty(t, b.String())
	})
}
