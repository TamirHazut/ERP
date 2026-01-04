package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	shared_models "erp.localhost/internal/infra/model/shared"
	"github.com/rs/zerolog"
)

type Logger struct {
	logger zerolog.Logger
}

// NewLogger creates a new logger for each module.
func NewLogger(module shared_models.Module) *Logger {
	baseLogger := zerolog.New(newConsoleWriter()).
		With().
		Timestamp().
		Str("module", string(module)).
		Logger()

	return &Logger{
		logger: baseLogger,
	}
}

// pipeFormatter is a custom writer that formats zerolog JSON output with pipe separators
type pipeFormatter struct {
	w io.Writer
}

func (f *pipeFormatter) Write(p []byte) (n int, err error) {
	var logData map[string]interface{}
	if err := json.Unmarshal(p, &logData); err != nil {
		// If it's not JSON, write as-is
		return f.w.Write(p)
	}

	var buf bytes.Buffer

	// Extract and format timestamp
	if ts, ok := logData[zerolog.TimestampFieldName]; ok {
		buf.WriteString(fmt.Sprintf("%v", ts))
		delete(logData, zerolog.TimestampFieldName)
	}
	buf.WriteString(" | ")

	// Extract and format level
	if level, ok := logData[zerolog.LevelFieldName]; ok {
		buf.WriteString(strings.ToUpper(fmt.Sprintf("%v", level)))
		delete(logData, zerolog.LevelFieldName)
	}
	buf.WriteString(" | ")

	// Extract and format module
	if module, ok := logData["module"]; ok {
		buf.WriteString(fmt.Sprintf("%v", module))
		delete(logData, "module")
	}
	buf.WriteString(" | ")

	// Extract and format message
	if msg, ok := logData[zerolog.MessageFieldName]; ok {
		buf.WriteString(fmt.Sprintf("%v", msg))
		delete(logData, zerolog.MessageFieldName)
	}

	// Add remaining fields in key=value format
	for k, v := range logData {
		buf.WriteString(" | ")
		buf.WriteString(k)
		buf.WriteString("=")

		// Format the value
		if str, ok := v.(string); ok {
			buf.WriteString(str)
		} else {
			buf.WriteString(fmt.Sprint(v))
		}
	}

	buf.WriteString("\n")
	_, err = f.w.Write(buf.Bytes())
	return len(p), err
}

// newConsoleWriter creates a console writer for the logger with custom pipe-separated format.
func newConsoleWriter() io.Writer {
	return &pipeFormatter{w: os.Stdout}
}

func (l *Logger) Trace(msg string, extraFields ...any) {
	l.log(zerolog.TraceLevel, msg, extraFields...)
}

func (l *Logger) Debug(msg string, extraFields ...any) {
	l.log(zerolog.DebugLevel, msg, extraFields...)
}

func (l *Logger) Info(msg string, extraFields ...any) {
	l.log(zerolog.InfoLevel, msg, extraFields...)
}

func (l *Logger) Warn(msg string, extraFields ...any) {
	l.log(zerolog.WarnLevel, msg, extraFields...)
}

func (l *Logger) Error(msg string, extraFields ...any) {
	l.log(zerolog.ErrorLevel, msg, extraFields...)
}

func (l *Logger) Fatal(msg string, extraFields ...any) {
	l.log(zerolog.FatalLevel, msg, extraFields...)
}

func (l *Logger) log(level zerolog.Level, msg string, extraFields ...any) {
	if len(extraFields)%2 != 0 {
		l.logger.Error().Msg("extraFields must be key-value pairs")
		return
	}

	ev := l.logger.WithLevel(level)

	for i := 0; i < len(extraFields); i += 2 {
		key, ok := extraFields[i].(string)
		if !ok {
			l.logger.Error().Msg("field keys must be strings")
			return
		}
		ev = ev.Interface(key, extraFields[i+1])
	}

	ev.Msg(msg)
}
