package logging

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
)

func newTestLogger(t *testing.T) (*logger, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}

	l := logrus.New()
	l.SetOutput(buf)
	l.SetLevel(logrus.DebugLevel)

	l.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: ISO8601,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})
	return &logger{entry: logrus.NewEntry(l)}, buf
}

func TestDebug(t *testing.T) {
	l, buf := newTestLogger(t)

	l.Debug("hello earth")
	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("output is not valid JSON: %v\nraw: %s", err, buf.String())
	}
	if entry["message"] != "hello earth" {
		t.Errorf("message = %v", entry["message"])
	}
}

func TestDebugf(t *testing.T) {
	l, buf := newTestLogger(t)

	l.Debugf("Hello %s", "earth")
	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("output is not valid JSON: %v\nraw: %s", err, buf.String())
	}
	if entry["message"] != "Hello earth" {
		t.Errorf("message = %v", entry["message"])
	}
}
