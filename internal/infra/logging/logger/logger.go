package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"erp.localhost/internal/infra/model/shared"
	"github.com/rs/zerolog"
)

//go:generate mockgen -destination=mock/mock_logger.go -package=mock erp.localhost/internal/infra/logging/logger Logger

type Logger interface {
	Trace(msg string, extraFields ...any)
	Debug(msg string, extraFields ...any)
	Info(msg string, extraFields ...any)
	Warn(msg string, extraFields ...any)
	Error(msg string, extraFields ...any)
	Fatal(msg string, extraFields ...any)
}

// FileOpenMode defines how log files should be opened
type FileOpenMode int

const (
	FileModeAppend FileOpenMode = iota
	FileModeTruncate
)

// LoggerConfig holds configuration for the logger from environment variables
type LoggerConfig struct {
	LogsDir        string
	FileMode       FileOpenMode
	ConsoleEnabled bool
	Module         shared.Module
}

type BaseLogger struct {
	logger      zerolog.Logger
	fileCleanup func()
}

// getLoggerConfigFromEnv reads logger configuration from environment variables
func getLoggerConfigFromEnv() LoggerConfig {
	config := LoggerConfig{
		LogsDir:        os.Getenv("LOG_FILE_PATH"),
		FileMode:       FileModeTruncate,
		ConsoleEnabled: true,
	}

	// Parse LOG_FILE_MODE
	if mode := os.Getenv("LOG_FILE_MODE"); mode == "append" {
		config.FileMode = FileModeAppend
	}

	// Parse LOG_CONSOLE_ENABLED
	if console := os.Getenv("LOG_CONSOLE_ENABLED"); console == "false" {
		config.ConsoleEnabled = false
	}

	return config
}

// findProjectRoot walks up the directory tree to find go.mod
func findProjectRoot() (string, error) {
	// Start from executable directory
	execPath, err := os.Executable()
	if err != nil {
		// Fallback to working directory
		return os.Getwd()
	}
	dir := filepath.Dir(execPath)

	// Walk up directory tree looking for go.mod
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding go.mod
			break
		}
		dir = parent
	}

	// Fallback to working directory
	return os.Getwd()
}

// createFileWriter creates a file writer with the specified configuration
func createFileWriter(config LoggerConfig) (io.Writer, *os.File, error) {
	if config.LogsDir == "" {
		return nil, nil, fmt.Errorf("file path is empty")
	}

	// Resolve file path
	filePath := fmt.Sprintf("%s/%s.log", config.LogsDir, strings.ToLower(string(config.Module)))
	if !filepath.IsAbs(filePath) {
		// Relative path - resolve from project root
		projectRoot, err := findProjectRoot()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to find project root: %w", err)
		}
		filePath = filepath.Join(projectRoot, filePath)
	}

	// Create directory if needed
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Determine file opening flags
	flags := os.O_CREATE | os.O_WRONLY
	if config.FileMode == FileModeAppend {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	// Open file
	file, err := os.OpenFile(filePath, flags, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Wrap in pipeFormatter
	writer := &pipeFormatter{w: file}
	return writer, file, nil
}

// createMultiWriter creates a writer based on configuration (console, file, or both)
func createMultiWriter(config LoggerConfig) (io.Writer, func(), error) {
	var writers []io.Writer
	var fileHandle *os.File

	// Cleanup function to close file handle
	cleanup := func() {
		if fileHandle != nil {
			fileHandle.Close()
		}
	}

	// Add console writer if enabled
	if config.ConsoleEnabled {
		writers = append(writers, &pipeFormatter{w: os.Stdout})
	}

	// Add file writer if path is provided
	if config.LogsDir != "" {
		fileWriter, file, err := createFileWriter(config)
		if err != nil {
			return nil, cleanup, err
		}
		writers = append(writers, fileWriter)
		fileHandle = file
	}

	// Return appropriate writer
	if len(writers) == 0 {
		// No writers - fallback to console
		return &pipeFormatter{w: os.Stdout}, cleanup, nil
	} else if len(writers) == 1 {
		return writers[0], cleanup, nil
	} else {
		return io.MultiWriter(writers...), cleanup, nil
	}
}

// NewBaseLogger creates a new logger for each module.
func NewBaseLogger(module shared.Module, opts ...map[string]string) *BaseLogger {
	// Read configuration from environment
	config := getLoggerConfigFromEnv()
	config.Module = module

	// Create multi-writer (console, file, or both)
	writer, cleanup, err := createMultiWriter(config)
	if err != nil {
		// If file logging fails, fall back to console-only and log warning
		writer = newConsoleWriter()
		cleanup = func() {}

		// Create a temporary logger to log the warning
		tempLogger := zerolog.New(writer).
			With().
			Timestamp().
			Str("module", string(module)).
			Logger()
		tempLogger.Warn().
			Err(err).
			Str("file_path", config.LogsDir).
			Msg("Failed to initialize file logging, using console only")
	}

	baseLogger := zerolog.New(writer).
		With().
		Timestamp().
		Str("module", string(module)).
		Logger()

	return &BaseLogger{
		logger:      baseLogger,
		fileCleanup: cleanup,
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

func (l *BaseLogger) Trace(msg string, extraFields ...any) {
	if l == nil {
		return
	}
	l.log(zerolog.TraceLevel, msg, extraFields...)
}

func (l *BaseLogger) Debug(msg string, extraFields ...any) {
	if l == nil {
		return
	}
	l.log(zerolog.DebugLevel, msg, extraFields...)
}

func (l *BaseLogger) Info(msg string, extraFields ...any) {
	if l == nil {
		return
	}
	l.log(zerolog.InfoLevel, msg, extraFields...)
}

func (l *BaseLogger) Warn(msg string, extraFields ...any) {
	if l == nil {
		return
	}
	l.log(zerolog.WarnLevel, msg, extraFields...)
}

func (l *BaseLogger) Error(msg string, extraFields ...any) {
	if l == nil {
		return
	}
	l.log(zerolog.ErrorLevel, msg, extraFields...)
}

func (l *BaseLogger) Fatal(msg string, extraFields ...any) {
	if l == nil {
		return
	}
	l.log(zerolog.FatalLevel, msg, extraFields...)
}

func (l *BaseLogger) log(level zerolog.Level, msg string, extraFields ...any) {
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

// Close releases any resources held by the logger (e.g., file handles)
func (l *BaseLogger) Close() error {
	if l != nil && l.fileCleanup != nil {
		l.fileCleanup()
	}
	return nil
}
