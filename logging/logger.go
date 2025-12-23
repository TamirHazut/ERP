package logging

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type Module string

const (
	ModuleAuth    Module = "Auth"
	ModuleConfig  Module = "Config"
	ModuleCore    Module = "Core"
	ModuleDB      Module = "DB"
	ModuleEvents  Module = "Events"
	ModuleGateway Module = "Gateway"
	ModuleSidecar Module = "Sidecar"
	ModuleWebUI   Module = "WebUI"
)

type Logger struct {
	logger zerolog.Logger
}

// NewLogger creates a new logger for each module.
func NewLogger(module Module) *Logger {
	baseLogger := zerolog.New(newConsoleWriter(module)).
		With().
		Timestamp().
		Str("module", string(module)).
		Logger()

	return &Logger{
		logger: baseLogger,
	}
}

// newConsoleWriter creates a console writer for the logger.
func newConsoleWriter(module Module) zerolog.ConsoleWriter {
	w := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	}

	w.FormatTimestamp = func(i interface{}) string {
		return i.(string)
	}

	w.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(i.(string))
	}

	w.FormatMessage = func(i interface{}) string {
		return i.(string)
	}

	w.FormatFieldName = func(i interface{}) string { return "" }
	w.FormatFieldValue = func(i interface{}) string { return i.(string) }

	w.FormatExtra = func(m map[string]interface{}, b *bytes.Buffer) error {
		// module first
		if v, ok := m["module"]; ok {
			b.WriteString(" | ")
			b.WriteString(v.(string))
			delete(m, "module")
		}

		for k, v := range m {
			b.WriteString(" | ")
			b.WriteString(k)
			b.WriteString("=")
			b.WriteString(fmt.Sprint(v))
		}

		return nil
	}

	return w
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
