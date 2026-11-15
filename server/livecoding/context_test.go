package livecoding

import (
	"context"
	"errors"
	"testing"
	"time"
)

// §4.4 — interruptible operation: the canonical select skeleton.
// The delay is injected so BOTH branches are testable fast; the
// "realistic" version is fetchReport(ctx) with a fixed 2s timer.
func fetchReportWithDelay(ctx context.Context, delay time.Duration) (string, error) {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return "report content", nil // work finished in time
	case <-ctx.Done():
		return "", ctx.Err() // caller gave up — THIS is the error branch
	}
}

func TestFetchReportCancelledByTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel() // always: go vet flags leaked cancels

	_, err := fetchReportWithDelay(ctx, 2*time.Second)

	if err == nil {
		t.Fatal("expected cancellation error, got none")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want context.DeadlineExceeded", err)
	}
}

func TestFetchReportCompletesWhenGivenTime(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	report, err := fetchReportWithDelay(ctx, 10*time.Millisecond)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report != "report content" {
		t.Errorf("report = %q", report)
	}
}

// DeadlineExceeded (parent had a deadline) vs Canceled (explicit cancel).
func TestCanceledVsDeadlineExceeded(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // explicit cancel, no deadline

	_, err := fetchReportWithDelay(ctx, time.Second)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("explicit cancel: err = %v, want context.Canceled", err)
	}
}

// §3 bonus — cancellation propagates to CHILD contexts.
func TestCancellationPropagates(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	child, cancelChild := context.WithTimeout(parent, 10*time.Second)
	defer cancelChild() // still must call: child's cancel is independent

	cancelParent()

	<-child.Done() // child observes the parent's cancellation
	if !errors.Is(child.Err(), context.Canceled) {
		t.Errorf("child err = %v, want context.Canceled (inherited)", child.Err())
	}
}
