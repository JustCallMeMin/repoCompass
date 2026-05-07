// Package rcerr provides structured error types with machine-readable error codes
// for the repoCompass local scan flow.
package rcerr

import (
	"errors"
	"fmt"
)

// ErrorCode is a machine-readable identifier for a category of error.
type ErrorCode string

const (
	// CodeInvalidSource is returned when the scan source is invalid,
	// e.g., empty path, path does not exist, or path is not a directory.
	CodeInvalidSource ErrorCode = "INVALID_SOURCE"

	// CodeConfigResolveFailed is returned when configuration resolution fails,
	// e.g., a .repocompass.yaml file exists but contains invalid YAML.
	CodeConfigResolveFailed ErrorCode = "CONFIG_RESOLVE_FAILED"

	// CodeRepoResolveFailed is returned when repository resolution fails
	// due to an operational error (filesystem, permissions, etc.).
	CodeRepoResolveFailed ErrorCode = "REPO_RESOLVE_FAILED"

	// CodeSnapshotCreateFailed is returned when snapshot creation fails.
	CodeSnapshotCreateFailed ErrorCode = "SNAPSHOT_CREATE_FAILED"

	// CodeScanExecutionFailed is returned for general scan orchestration failures.
	CodeScanExecutionFailed ErrorCode = "SCAN_EXECUTION_FAILED"

	// CodeInternalError is returned for unexpected internal failures.
	CodeInternalError ErrorCode = "INTERNAL_ERROR"

	// CodeDatabaseUnavailable is returned when PostgreSQL cannot be reached.
	CodeDatabaseUnavailable ErrorCode = "DATABASE_UNAVAILABLE"

	// CodeDatabaseConstraintFailed is returned when PostgreSQL rejects data due
	// to a schema constraint such as a foreign key, unique, or check violation.
	CodeDatabaseConstraintFailed ErrorCode = "DATABASE_CONSTRAINT_FAILED"

	// CodeDatabaseQueryFailed is returned for database read/write failures that
	// are not classified more specifically.
	CodeDatabaseQueryFailed ErrorCode = "DATABASE_QUERY_FAILED"

	// CodeDatabaseNotFound is returned when a requested persisted row is absent.
	CodeDatabaseNotFound ErrorCode = "DATABASE_NOT_FOUND"

	// CodeDatabaseTimeout is returned when a database operation times out or is
	// cancelled through context propagation.
	CodeDatabaseTimeout ErrorCode = "DATABASE_TIMEOUT"
)

// Error is a structured error carrying a machine-readable ErrorCode, a
// human-readable message, and an optional underlying cause.
type Error struct {
	Code    ErrorCode
	Message string
	Err     error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error so that errors.Is and errors.As work
// transparently through rcerr.Error wrapping.
func (e *Error) Unwrap() error {
	return e.Err
}

// New creates a new *Error with the given code, human-readable message, and
// optional underlying cause. cause may be nil.
func New(code ErrorCode, message string, cause error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     cause,
	}
}

// CodeOf extracts the ErrorCode from err if it is (or wraps) an *rcerr.Error.
// Returns the code and true on success, or ("", false) if no *rcerr.Error is
// found in the error chain.
func CodeOf(err error) (ErrorCode, bool) {
	var rcErr *Error
	if errors.As(err, &rcErr) {
		return rcErr.Code, true
	}
	return "", false
}
