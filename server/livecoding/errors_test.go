package livecoding

import (
	"errors"
	"fmt"
	"testing"
)

// §4.2 — sentinel error: a package-level value callers can compare with errors.Is.
var ErrNotFound = errors.New("not found")

// §4.2 — typed error: carries structured data, extracted with errors.As.
type NotFoundError struct {
	Key string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("not found: %s", e.Key)
}

func TestSentinelErrorWithIs(t *testing.T) {
	// %w wraps: the chain preserves identity for errors.Is.
	wrapped := fmt.Errorf("lookup city: %w", ErrNotFound)

	if !errors.Is(wrapped, ErrNotFound) {
		t.Errorf("errors.Is failed through wrapping: %v", wrapped)
	}

	// Unrelated sentinel must NOT match.
	if errors.Is(wrapped, errors.New("other")) {
		t.Error("errors.Is matched an unrelated error")
	}
}

func TestTypedErrorWithAs(t *testing.T) {
	wrapped := fmt.Errorf("handler: %w", &NotFoundError{Key: "city:1"})

	var nf *NotFoundError
	if !errors.As(wrapped, &nf) {
		t.Fatalf("errors.As failed: %v", wrapped)
	}
	if nf.Key != "city:1" {
		t.Errorf("Key = %q, want %q", nf.Key, "city:1")
	}
}

func TestIsVsAsDistinction(t *testing.T) {
	typed := fmt.Errorf("handler: %w", &NotFoundError{Key: "x"})

	// Is matches VALUES (sentinels), As matches TYPES. They are not interchangeable.
	if errors.Is(typed, ErrNotFound) {
		t.Error("sentinel should not match a different, structurally similar type")
	}
}
