package github

import "time"

// Integration links one GitHub repository to one RepoCompass repository.
type Integration struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	RepositoryID   string    `json:"repository_id"`
	Owner          string    `json:"owner"`
	Repo           string    `json:"repo"`
	CloneURL       string    `json:"clone_url"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// WebhookEvent is one accepted or ignored GitHub webhook delivery.
type WebhookEvent struct {
	ID                 string    `json:"id"`
	DeliveryID         string    `json:"delivery_id"`
	EventType          string    `json:"event_type"`
	RepositoryFullName string    `json:"repository_full_name"`
	RepositoryCloneURL string    `json:"repository_clone_url"`
	Payload            []byte    `json:"payload"`
	Status             string    `json:"status"`
	ErrorMessage       string    `json:"error_message,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

// ScanJob is an async scan execution request created by API or webhook intake.
type ScanJob struct {
	ID             string     `json:"id"`
	RepositoryID   string     `json:"repository_id"`
	SourceType     string     `json:"source_type"`
	SourceURL      string     `json:"source_url,omitempty"`
	Status         string     `json:"status"`
	ScanID         string     `json:"scan_id,omitempty"`
	WebhookEventID string     `json:"webhook_event_id,omitempty"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

const (
	// WebhookStatusAccepted marks a webhook that can trigger follow-up work.
	WebhookStatusAccepted = "accepted"
	// WebhookStatusIgnored marks a webhook that was valid but irrelevant.
	WebhookStatusIgnored = "ignored"
	// WebhookStatusFailed marks a webhook that failed after validation.
	WebhookStatusFailed = "failed"

	// JobStatusQueued means a job is waiting for the worker.
	JobStatusQueued = "queued"
	// JobStatusRunning means a job is executing.
	JobStatusRunning = "running"
	// JobStatusCompleted means a job finished and persisted a scan.
	JobStatusCompleted = "completed"
	// JobStatusFailed means a job failed with a stored error.
	JobStatusFailed = "failed"
)
