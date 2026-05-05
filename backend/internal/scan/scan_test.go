package scan_test

import (
	"errors"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
)

func TestScan_TransitionTo_Valid(t *testing.T) {
	tests := []struct {
		name   string
		start  scan.Status
		target scan.Status
	}{
		// From Created
		{"Created to Queued", scan.StatusCreated, scan.StatusQueued},
		{"Created to Running", scan.StatusCreated, scan.StatusRunning},
		{"Created to Cancelled", scan.StatusCreated, scan.StatusCancelled},

		// From Queued
		{"Queued to Running", scan.StatusQueued, scan.StatusRunning},
		{"Queued to Cancelled", scan.StatusQueued, scan.StatusCancelled},

		// From Running
		{"Running to Completed", scan.StatusRunning, scan.StatusCompleted},
		{"Running to Failed", scan.StatusRunning, scan.StatusFailed},
		{"Running to Cancelled", scan.StatusRunning, scan.StatusCancelled},

		// Uninitialized
		{"Empty to Created", "", scan.StatusCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &scan.Scan{
				Status: tt.start,
			}
			err := s.TransitionTo(tt.target)
			if err != nil {
				t.Fatalf("expected transition from %q to %q to succeed, got error: %v", tt.start, tt.target, err)
			}
			if s.Status != tt.target {
				t.Errorf("expected status to be %q, got %q", tt.target, s.Status)
			}
		})
	}
}

func TestScan_TransitionTo_Invalid(t *testing.T) {
	tests := []struct {
		name   string
		start  scan.Status
		target scan.Status
	}{
		// From Created
		{"Created to Completed", scan.StatusCreated, scan.StatusCompleted},
		{"Created to Failed", scan.StatusCreated, scan.StatusFailed},

		// From Queued
		{"Queued to Completed", scan.StatusQueued, scan.StatusCompleted},
		{"Queued to Failed", scan.StatusQueued, scan.StatusFailed},
		{"Queued to Created", scan.StatusQueued, scan.StatusCreated},

		// From Running
		{"Running to Created", scan.StatusRunning, scan.StatusCreated},
		{"Running to Queued", scan.StatusRunning, scan.StatusQueued},

		// From Terminal States
		{"Completed to Running", scan.StatusCompleted, scan.StatusRunning},
		{"Completed to Failed", scan.StatusCompleted, scan.StatusFailed},
		{"Completed to Completed", scan.StatusCompleted, scan.StatusCompleted}, // even self-transition is blocked from terminal

		{"Failed to Running", scan.StatusFailed, scan.StatusRunning},
		{"Failed to Completed", scan.StatusFailed, scan.StatusCompleted},

		{"Cancelled to Running", scan.StatusCancelled, scan.StatusRunning},
		{"Cancelled to Completed", scan.StatusCancelled, scan.StatusCompleted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &scan.Scan{
				Status: tt.start,
			}
			err := s.TransitionTo(tt.target)
			if err == nil {
				t.Fatalf("expected transition from %q to %q to fail, but it succeeded", tt.start, tt.target)
			}

			if !errors.Is(err, scan.ErrInvalidStateTransition) {
				t.Errorf("expected error to wrap ErrInvalidStateTransition, got %v", err)
			}

			if s.Status != tt.start {
				t.Errorf("expected status to remain %q, got %q", tt.start, s.Status)
			}
		})
	}
}
