package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger is a wrapper around zerolog.Logger
type Logger struct {
	zerolog.Logger
}

// NewLogger creates a new Logger instance from a zerolog.Logger
func NewLogger(logger zerolog.Logger) *Logger {
	return &Logger{Logger: logger}
}

var (
	// defaultLogger is the default logger instance
	defaultLogger zerolog.Logger
)

// LogLevel represents the log level
type LogLevel string

const (
	// DebugLevel is for debug messages
	DebugLevel LogLevel = "debug"
	// InfoLevel is for informational messages
	InfoLevel LogLevel = "info"
	// WarnLevel is for warning messages
	WarnLevel LogLevel = "warn"
	// ErrorLevel is for error messages
	ErrorLevel LogLevel = "error"
	// FatalLevel is for fatal messages (panics after logging)
	FatalLevel LogLevel = "fatal"
)

// Config represents logger configuration
type Config struct {
	// Level is the log level
	Level LogLevel
	// Pretty enables pretty logging (human-readable format)
	Pretty bool
	// Output is the output writer (defaults to os.Stdout)
	Output io.Writer
}

// Configure configures the logger with the provided config
func Configure(config Config) {
	// Default values
	if config.Output == nil {
		config.Output = os.Stdout
	}

	// Set global time format
	zerolog.TimeFieldFormat = time.RFC3339

	// Set global log level
	switch config.Level {
	case DebugLevel:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case InfoLevel:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case WarnLevel:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case ErrorLevel:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case FatalLevel:
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Create writer
	var writer io.Writer = config.Output
	if config.Pretty {
		writer = zerolog.ConsoleWriter{
			Out:        config.Output,
			TimeFormat: time.RFC3339,
		}
	}

	// Set default logger
	defaultLogger = zerolog.New(writer).With().Timestamp().Logger()
	log.Logger = defaultLogger
}

// Debug logs a debug message
func Debug() *zerolog.Event {
	return defaultLogger.Debug()
}

// Info logs an informational message
func Info() *zerolog.Event {
	return defaultLogger.Info()
}

// Warn logs a warning message
func Warn() *zerolog.Event {
	return defaultLogger.Warn()
}

// Error logs an error message
func Error() *zerolog.Event {
	return defaultLogger.Error()
}

// Fatal logs a fatal message and then panics
func Fatal() *zerolog.Event {
	return defaultLogger.Fatal()
}

// WithField adds a field to the logger
func WithField(key string, value interface{}) zerolog.Logger {
	return defaultLogger.With().Interface(key, value).Logger()
}

// WithFields adds multiple fields to the logger
func WithFields(fields map[string]interface{}) zerolog.Logger {
	context := defaultLogger.With()
	for k, v := range fields {
		context = context.Interface(k, v)
	}
	return context.Logger()
}

// init initializes the default logger
func init() {
	// Default configuration
	Configure(Config{
		Level:  InfoLevel,
		Pretty: true,
		Output: os.Stdout,
	})
}
