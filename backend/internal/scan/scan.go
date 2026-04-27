// Package scan contains scan orchestration and scan execution primitives.
package scan

import (
	"errors"
	"fmt"
	"time"
)

// Status represents the lifecycle state of a scan.
type Status string

const (
	// StatusCreated means a scan object exists, but execution has not been scheduled or started.
	StatusCreated Status = "created"
	// StatusQueued means a scan is accepted for execution and waiting to run.
	StatusQueued Status = "queued"
	// StatusRunning means repository resolution, snapshot creation, or analyzer execution is in progress.
	StatusRunning Status = "running"
	// StatusCompleted means the scan finished successfully and produced a result.
	StatusCompleted Status = "completed"
	// StatusFailed means the scan stopped because of a recoverable or user-facing error.
	StatusFailed Status = "failed"
	// StatusCancelled means the scan was intentionally stopped before completion.
	StatusCancelled Status = "cancelled"
)

// Scan represents one analysis run over one repository snapshot.
type Scan struct {
	ID           string
	SnapshotID   string
	Status       Status
	StartTime    *time.Time
	EndTime      *time.Time
	ErrorDetails string
}

// ErrInvalidStateTransition is returned when a scan attempts an invalid status transition.
var ErrInvalidStateTransition = errors.New("invalid state transition")

// TransitionTo updates the scan's status to the new state if the transition is valid.
func (s *Scan) TransitionTo(newStatus Status) error {
	switch s.Status {
	case StatusCreated:
		if newStatus == StatusQueued || newStatus == StatusRunning || newStatus == StatusCancelled {
			s.Status = newStatus
			return nil
		}
	case StatusQueued:
		if newStatus == StatusRunning || newStatus == StatusCancelled {
			s.Status = newStatus
			return nil
		}
	case StatusRunning:
		if newStatus == StatusCompleted || newStatus == StatusFailed || newStatus == StatusCancelled {
			s.Status = newStatus
			return nil
		}
	case StatusCompleted, StatusFailed, StatusCancelled:
		// Terminal states cannot transition to anything else, not even themselves (ideally should be an error to try to transition from completed -> completed)
		return fmt.Errorf("%w: cannot transition from terminal state %q to %q", ErrInvalidStateTransition, s.Status, newStatus)
	default:
		// Unknown state or initial empty state (treat empty as uninitialized)
		if s.Status == "" && newStatus == StatusCreated {
			s.Status = newStatus
			return nil
		}
		return fmt.Errorf("%w: unknown current state %q", ErrInvalidStateTransition, s.Status)
	}

	return fmt.Errorf("%w: cannot transition from %q to %q", ErrInvalidStateTransition, s.Status, newStatus)
}
