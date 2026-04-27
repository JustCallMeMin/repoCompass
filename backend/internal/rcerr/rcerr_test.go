package rcerr_test

import (
	"errors"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/rcerr"
)

func TestError_Error_WithCause(t *testing.T) {
	cause := errors.New("underlying failure")
	err := rcerr.New(rcerr.CodeInvalidSource, "path does not exist", cause)

	got := err.Error()
	if got == "" {
		t.Fatal("expected non-empty error string")
	}
	// Should contain the code, message, and cause
	if !contains(got, string(rcerr.CodeInvalidSource)) {
		t.Errorf("error string %q does not contain code %q", got, rcerr.CodeInvalidSource)
	}
	if !contains(got, "path does not exist") {
		t.Errorf("error string %q does not contain message", got)
	}
	if !contains(got, "underlying failure") {
		t.Errorf("error string %q does not contain cause", got)
	}
}

func TestError_Error_WithoutCause(t *testing.T) {
	err := rcerr.New(rcerr.CodeConfigResolveFailed, "invalid YAML", nil)

	got := err.Error()
	if !contains(got, string(rcerr.CodeConfigResolveFailed)) {
		t.Errorf("error string %q does not contain code", got)
	}
	if !contains(got, "invalid YAML") {
		t.Errorf("error string %q does not contain message", got)
	}
}

func TestError_Unwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := rcerr.New(rcerr.CodeRepoResolveFailed, "resolution failed", cause)

	if !errors.Is(err, cause) {
		t.Error("errors.Is should find the cause through Unwrap")
	}
}

func TestCodeOf_RcErr(t *testing.T) {
	err := rcerr.New(rcerr.CodeSnapshotCreateFailed, "create failed", nil)

	code, ok := rcerr.CodeOf(err)
	if !ok {
		t.Fatal("expected CodeOf to succeed for *rcerr.Error")
	}
	if code != rcerr.CodeSnapshotCreateFailed {
		t.Errorf("expected code %q, got %q", rcerr.CodeSnapshotCreateFailed, code)
	}
}

func TestCodeOf_WrappedRcErr(t *testing.T) {
	inner := rcerr.New(rcerr.CodeInvalidSource, "bad path", nil)
	outer := errors.New("outer: " + inner.Error())
	// Wrap via fmt.Errorf %w equivalent using errors wrapping
	wrapped := &wrappedErr{msg: "outer", cause: inner}

	code, ok := rcerr.CodeOf(wrapped)
	if !ok {
		t.Fatal("expected CodeOf to find rcerr.Error in chain")
	}
	if code != rcerr.CodeInvalidSource {
		t.Errorf("expected code %q, got %q", rcerr.CodeInvalidSource, code)
	}
	_ = outer // suppress unused warning
}

func TestCodeOf_NonRcErr(t *testing.T) {
	err := errors.New("plain error")

	_, ok := rcerr.CodeOf(err)
	if ok {
		t.Error("expected CodeOf to return false for plain error")
	}
}

// wrappedErr is a helper to simulate a wrapping error in tests.
type wrappedErr struct {
	msg   string
	cause error
}

func (w *wrappedErr) Error() string { return w.msg + ": " + w.cause.Error() }
func (w *wrappedErr) Unwrap() error { return w.cause }

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
