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
	if out := buf.String(); !strings.Contains(out, `[http] GET /api/v1/cities`) ||
		!strings.Contains(out, `418`) {
		t.Errorf("monitor line missing, got: %s", out)
	}
}

func TestRequestMiddlewareSetsRequestID(t *testing.T) {
	var called bool
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if logging.FromContext(r.Context()) == nil {
			t.Error("no logger in request context")
		}
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/ping", nil)

	handlers.RequestMiddleware(testLoggerHelper(&bytes.Buffer{}))(next).ServeHTTP(rec, req)

	if !called {
		t.Error("next handler was not called")
	}
	if id := rec.Header().Get("X-Request-ID"); id == "" {
		t.Error("X-Request-ID header missing")
	}
}

func TestMiddlewareChainCorrelatesRequests(t *testing.T) {
	var buf bytes.Buffer
	log := testLoggerHelper(&buf)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// RequestMiddleware outermost — same order as RunHTTP
	chain := handlers.RequestMiddleware(log)(handlers.MonitorMiddleware(log)(inner))

	rec1, rec2 := httptest.NewRecorder(), httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/api/v1/cities", nil)
	req2 := httptest.NewRequest("GET", "/api/v1/cities", nil)
	chain.ServeHTTP(rec1, req1)
	chain.ServeHTTP(rec2, req2)

	id1, id2 := rec1.Header().Get("X-Request-ID"), rec2.Header().Get("X-Request-ID")
	if id1 == "" || id2 == "" {
		t.Fatal("X-Request-ID missing on a response")
	}
	if id1 == id2 {
		t.Errorf("two requests share one request_id: %s", id1)
	}

	// every monitor line must carry the request_id field
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if !strings.Contains(line, `"request_id"`) {
			t.Errorf("monitor line missing request_id: %s", line)
		}
	}
}
