package logging

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	// ISO8601 is timestamp format used by Logger.
	ISO8601        = "2006-01-02T15:04:05.000Z07:00"
	loggerFieldKey = "logger"
)

var exit = os.Exit

type Logger interface {
	Debug(...interface{})
	Debugf(string, ...interface{})

	Info(...interface{})
	Infof(string, ...interface{})

	Warn(...interface{})
	Warnf(string, ...interface{})

	Error(...interface{})
	Errorf(string, ...interface{})

	Fatal(...interface{})
	Fatalf(string, ...interface{})

	With(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

type logger struct {
	entry *logrus.Entry
	depth int
}

func (l logger) Debug(args ...interface{}) {
	l.sourced(l.depth).Debug(args...)
}

func (l logger) Debugf(format string, args ...interface{}) {
	l.sourced(l.depth).Debugf(format, args...)
}

func (l logger) Info(args ...interface{}) {
	l.sourced(l.depth).Info(args...)
}

func (l logger) Infof(format string, args ...interface{}) {
	l.sourced(l.depth).Infof(format, args...)
}

func (l logger) Warn(args ...interface{}) {
	l.sourced(l.depth).Warn(args...)
}

func (l logger) Warnf(format string, args ...interface{}) {
	l.sourced(l.depth).Warnf(format, args...)
}

func (l logger) Error(args ...interface{}) {
	l.sourced(l.depth).Error(args...)
}

func (l logger) Errorf(format string, args ...interface{}) {
	l.sourced(l.depth).Errorf(format, args...)
}

func (l logger) Fatal(args ...interface{}) {
	l.depth++
	l.Error(args...)
	exit(1)
}

func (l logger) Fatalf(format string, args ...interface{}) {
	l.depth++
	l.Errorf(format, args...)
	exit(1)
}

// With adds kv pair to log message.
func (l logger) With(key string, value interface{}) Logger {
	return logger{entry: l.entry.WithField(key, value)}
}

// WithFields adds map as a kv pairs to log message.
func (l logger) WithFields(fields map[string]interface{}) Logger {
	return logger{entry: l.entry.WithFields(fields)}
}

func (l logger) sourced(depth int) *logrus.Entry {
	_, file, line, ok := runtime.Caller(depth + 2)
	if !ok {
		file = "<???>"
		line = 1
	} else if slash := strings.LastIndex(file, "/"); slash >= 0 {
		file = file[slash+1:]
	}
	return l.entry.WithField(loggerFieldKey, fmt.Sprintf("%s:%d", file, line))
}

// Logger will automatically collect metrics (log event counters) for Prometheus.
// Metrics will be exposed only if you run metrics.ManagementServer in your application.
func parseConfig() (logLevel logrus.Level, outputFormat logrus.Formatter, err error) {
	// default setting
	logLevel = logrus.InfoLevel
	outputFormat = &logrus.JSONFormatter{}
	level := os.Getenv("LOGGING_LEVEL")
	switch strings.ToLower(level) {
	case "debug":
		logLevel = logrus.DebugLevel
	case "info", "":
		logLevel = logrus.InfoLevel
	case "warn":
		logLevel = logrus.WarnLevel
	case "error":
		logLevel = logrus.ErrorLevel
	case "fatal":
		logLevel = logrus.FatalLevel
		// note: any other value is silently treated as info for now —
		// documented deviation, see plan Phase 5 for aligning with
		// operations-execution's log-and-continue behavior.
	}
	format := os.Getenv("LOGGING_FORMAT")
	switch format {
	case "json", "": // empty means default
		outputFormat = &logrus.JSONFormatter{
			TimestampFormat: ISO8601,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "@timestamp",
				logrus.FieldKeyLevel: "@level",
				logrus.FieldKeyMsg:   "@message",
				logrus.FieldKeyFunc:  "@caller",
			},
		}
	case "text", "txt":
		outputFormat = &logrus.TextFormatter{FullTimestamp: true}
	default:
		return logLevel, outputFormat, errors.New("invalid logging format")
	}

	return logLevel, outputFormat, nil
}

func NewFromLogrus(l *logrus.Logger) Logger {
	return logger{entry: logrus.NewEntry(l)}
}

func NewLogger() Logger {
	logLevel, outputFormat, err := parseConfig()
	l := &logrus.Logger{
		Out:       os.Stderr,
		Formatter: outputFormat,
		Level:     logLevel,
	}
	pLogger := logger{entry: logrus.NewEntry(l)}
	if err != nil {
		// log through the wrapper so config errors get the same treatment
		pLogger.Errorf("Error parsing logger config: %s", err)
	}
	return pLogger
}

// loggerContextKey is the context key for a request-scoped Logger.
type loggerContextKey struct{}

// NewContext returns a context carrying the given logger.
func NewContext(ctx context.Context, log Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, log)
}

// FromContext returns the request-scoped logger stored in ctx, or nil.
func FromContext(ctx context.Context) Logger {
	if log, ok := ctx.Value(loggerContextKey{}).(Logger); ok {
		return log
	}
	return nil
}
