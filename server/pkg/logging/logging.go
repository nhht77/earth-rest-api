package logging

import (
	"errors"
	"os"

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
}

type logger struct {
	entry *logrus.Entry
	depth int
}

func (l logger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

func (l logger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

func (l logger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

func (l logger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

func (l logger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

func (l logger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

func (l logger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

func (l logger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

func (l logger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

func (l logger) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

// Logger will automatically collect metrics (log event counters) for Prometheus.
// Metrics will be exposed only if you run metrics.ManagementServer in your application.
func parseConfig() (logLevel logrus.Level, outputFormat logrus.Formatter, err error) {
	// default setting
	logLevel = logrus.InfoLevel
	outputFormat = &logrus.JSONFormatter{}
	level := os.Getenv("LOG_LEVEL")
	switch level {
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
	}
	loggingFormat := os.Getenv("LOGGING_FORMAT")
	switch loggingFormat {
	case "json":
		outputFormat = &logrus.JSONFormatter{
			TimestampFormat: ISO8601,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "@timestamp",
				logrus.FieldKeyLevel: "@level",
				logrus.FieldKeyMsg:   "@message",
				logrus.FieldKeyFunc:  "@caller",
			},
		}
	case "text":
		outputFormat = &logrus.TextFormatter{}
	default:
		return logLevel, outputFormat, errors.New("invalid logging format")
	}

	return logLevel, outputFormat, nil
}

func NewLogger() Logger {
	logLevel, outputFormat, err := parseConfig()
	l := &logrus.Logger{
		Out:       os.Stderr,
		Formatter: outputFormat,
		Level:     logLevel,
	}
	if err != nil {
		l.Error(err)
	}
	return logger{entry: logrus.NewEntry(l)}
}
