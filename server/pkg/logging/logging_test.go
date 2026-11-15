package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

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

func TestExample(t *testing.T) {
	// primitive array
	// copy of array
	// test mutation

	// struct array
	// copy of struct array
	// test mutation

	// struct
	// copy of struct
	// test mutation

}

func TestErrorChecker(t *testing.T) {
	//errors.new("type")
	// errors.is()
	//errors.Is()

}

func TestGoRoutineSameCounter(t *testing.T) {
	// primitive value
	// waitGroup
	// goroutine update source with/without protection
	// go func() {} ()

}

func fetchReport(ctx context.Context) (string, error) {
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()

	select {
	case <-timer.C:
		return "ready", nil
		// case <-ctx.Done():
		// 	return "", ctx.Err()
	}
}

func TestContextCancelAndDone(t *testing.T) {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		100*time.Millisecond,
	)
	defer cancel()

	//ctx := context.Background()

	// defer cancel()

	report, err := fetchReport(ctx)
	if err != nil {
		t.Fatalf("fetchReport failed: %v", err)
	}
	fmt.Println(report)

}

// explore the use case
// normally on TCP/UDP socket
func testFIFOWithBufferedChannel(t *testing.T) {
	ch := make(chan int, 10)
	ch <- 1
	ch <- 2
	ch <- 3
	ch <- 4
	ch <- 5
	ch <- 6
	close(ch)

	//process
	for v := range ch {
		fmt.Println(v)
	}
}

// Cache expireable for single/whole list

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
