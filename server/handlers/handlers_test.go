package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nhht77/earth-rest-api/server/handlers"
	"github.com/nhht77/earth-rest-api/server/pkg/logging"
	"github.com/sirupsen/logrus"
)

func testLoggerHelper(buf *bytes.Buffer) logging.Logger {
	l := logrus.New()
	l.SetOutput(buf)
	l.SetFormatter(&logrus.JSONFormatter{})
	return logging.NewFromLogrus(l)
}

func TestPing(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/ping", nil)

	handlers.Ping(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestMonitorMiddlewareLogsRequest(t *testing.T) {
	var buf bytes.Buffer
	log := testLoggerHelper(&buf)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot) // 418 — a status only tests ever see
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/cities", nil)

	handlers.MonitorMiddleware(log)(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want 418", rec.Code)
	}
	if out := buf.String(); !strings.Contains(out, `"[http] GET /api/v1/cities"`) {
		t.Errorf("monitor line missing, got: %s", out)
	}
}
