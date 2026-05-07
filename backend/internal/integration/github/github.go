// Package github contains GitHub integration helpers.
package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrInvalidSignature is returned when webhook signature validation fails.
var ErrInvalidSignature = errors.New("invalid github webhook signature")

// RepositoryURL identifies a public GitHub repository.
type RepositoryURL struct {
	Owner string
	Name  string
	URL   string
}

// Checkout is a temporary local checkout of a GitHub repository.
type Checkout struct {
	Path    string
	Cleanup func()
}

// PublicCloner clones public GitHub repositories with git.
type PublicCloner struct {
	BaseDir string
}

// ParseRepositoryURL parses https://github.com/{owner}/{repo} URLs.
func ParseRepositoryURL(raw string) (RepositoryURL, error) {
	if strings.TrimSpace(raw) == "" {
		return RepositoryURL{}, fmt.Errorf("github url is required")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return RepositoryURL{}, fmt.Errorf("parse github url: %w", err)
	}
	if parsed.Scheme != "https" || parsed.Host != "github.com" {
		return RepositoryURL{}, fmt.Errorf("github url must use https://github.com")
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return RepositoryURL{}, fmt.Errorf("github url must include owner and repository")
	}
	name := strings.TrimSuffix(parts[1], ".git")
	if name == "" {
		return RepositoryURL{}, fmt.Errorf("github repository name cannot be empty")
	}
	return RepositoryURL{
		Owner: parts[0],
		Name:  name,
		URL:   "https://github.com/" + parts[0] + "/" + name,
	}, nil
}

// Clone clones a public GitHub repository into a temporary local directory.
func (c PublicCloner) Clone(ctx context.Context, repo RepositoryURL) (Checkout, error) {
	baseDir := c.BaseDir
	if baseDir == "" {
		baseDir = os.TempDir()
	}
	parent, err := os.MkdirTemp(baseDir, "repocompass-github-*")
	if err != nil {
		return Checkout{}, fmt.Errorf("create github checkout temp dir: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(parent) }
	path := filepath.Join(parent, repo.Owner+"-"+repo.Name)
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", repo.URL, path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		cleanup()
		return Checkout{}, fmt.Errorf("git clone github repository: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return Checkout{Path: path, Cleanup: cleanup}, nil
}

// WebhookPayload is the minimal GitHub webhook payload shape used by M4.
type WebhookPayload struct {
	Repository struct {
		FullName string `json:"full_name"`
		CloneURL string `json:"clone_url"`
	} `json:"repository"`
}

// WebhookRequest contains validated webhook metadata and decoded payload.
type WebhookRequest struct {
	Event      string
	DeliveryID string
	Body       []byte
	Payload    WebhookPayload
}

// ReadWebhook validates and decodes a GitHub webhook request.
func ReadWebhook(r *http.Request, secret string) (string, WebhookPayload, error) {
	event := r.Header.Get("X-GitHub-Event")
	if event == "" {
		return "", WebhookPayload{}, fmt.Errorf("X-GitHub-Event is required")
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", WebhookPayload{}, fmt.Errorf("read webhook body: %w", err)
	}
	if secret != "" && !validSignature(body, secret, r.Header.Get("X-Hub-Signature-256")) {
		return "", WebhookPayload{}, ErrInvalidSignature
	}
	var payload WebhookPayload
	if len(body) > 0 {
		if err := json.Unmarshal(body, &payload); err != nil {
			return "", WebhookPayload{}, fmt.Errorf("decode webhook payload: %w", err)
		}
	}
	return event, payload, nil
}

// ReadWebhookRequest validates and decodes a GitHub webhook request with delivery metadata.
func ReadWebhookRequest(r *http.Request, secret string) (WebhookRequest, error) {
	event := r.Header.Get("X-GitHub-Event")
	if event == "" {
		return WebhookRequest{}, fmt.Errorf("X-GitHub-Event is required")
	}
	deliveryID := r.Header.Get("X-GitHub-Delivery")
	if deliveryID == "" {
		return WebhookRequest{}, fmt.Errorf("X-GitHub-Delivery is required")
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return WebhookRequest{}, fmt.Errorf("read webhook body: %w", err)
	}
	if secret != "" && !validSignature(body, secret, r.Header.Get("X-Hub-Signature-256")) {
		return WebhookRequest{}, ErrInvalidSignature
	}
	var payload WebhookPayload
	if len(body) > 0 {
		if err := json.Unmarshal(body, &payload); err != nil {
			return WebhookRequest{}, fmt.Errorf("decode webhook payload: %w", err)
		}
	}
	return WebhookRequest{Event: event, DeliveryID: deliveryID, Body: body, Payload: payload}, nil
}

func validSignature(body []byte, secret, header string) bool {
	if !strings.HasPrefix(header, "sha256=") {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(header))
}
