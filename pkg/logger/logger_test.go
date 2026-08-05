package logger

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestDebugDisabledByDefault(t *testing.T) {
	t.Setenv("LOG_LEVEL", "")
	var output bytes.Buffer
	previous := DebugLogger
	DebugLogger = log.New(&output, "", 0)
	t.Cleanup(func() { DebugLogger = previous })

	Debug("sentinel-secret")
	Debugf("value=%s", "sentinel-secret")
	if output.Len() != 0 {
		t.Fatalf("debug output with default log level = %q", output.String())
	}
}

func TestDebugEnabledExplicitly(t *testing.T) {
	t.Setenv("LOG_LEVEL", "debug")
	var output bytes.Buffer
	previous := DebugLogger
	DebugLogger = log.New(&output, "", 0)
	t.Cleanup(func() { DebugLogger = previous })

	Debugf("method=%s", "s3")
	if !strings.Contains(output.String(), "method=s3") {
		t.Fatalf("debug output = %q", output.String())
	}
}
