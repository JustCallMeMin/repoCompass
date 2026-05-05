package github

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseRepositoryURL(t *testing.T) {
	repo, err := ParseRepositoryURL("https://github.com/owner/repo.git")
	if err != nil {
		t.Fatalf("expected parse to succeed: %v", err)
	}
	if repo.Owner != "owner" || repo.Name != "repo" || repo.URL != "https://github.com/owner/repo" {
		t.Fatalf("unexpected repo: %#v", repo)
	}
}

func TestParseRepositoryURLRejectsInvalidURL(t *testing.T) {
	for _, raw := range []string{"", "https://example.com/owner/repo", "http://github.com/owner/repo", "https://github.com/owner"} {
		t.Run(raw, func(t *testing.T) {
			if _, err := ParseRepositoryURL(raw); err == nil {
				t.Fatal("expected invalid url to fail")
			}
		})
	}
}

func TestReadWebhookWithoutSecret(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString(`{"repository":{"full_name":"owner/repo"}}`))
	request.Header.Set("X-GitHub-Event", "push")

	event, payload, err := ReadWebhook(request, "")
	if err != nil {
		t.Fatalf("expected webhook read to succeed: %v", err)
	}
	if event != "push" {
		t.Fatalf("unexpected event: got %q", event)
	}
	if payload.Repository.FullName != "owner/repo" {
		t.Fatalf("unexpected repository: got %q", payload.Repository.FullName)
	}
}

func TestReadWebhookValidatesSignature(t *testing.T) {
	body := []byte(`{"repository":{"full_name":"owner/repo"}}`)
	request := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	request.Header.Set("X-GitHub-Event", "push")
	request.Header.Set("X-Hub-Signature-256", signature(body, "secret"))

	if _, _, err := ReadWebhook(request, "secret"); err != nil {
		t.Fatalf("expected signature to validate: %v", err)
	}
}

func TestReadWebhookRejectsInvalidSignature(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString(`{}`))
	request.Header.Set("X-GitHub-Event", "push")
	request.Header.Set("X-Hub-Signature-256", "sha256=bad")

	_, _, err := ReadWebhook(request, "secret")
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected invalid signature error, got %v", err)
	}
}

func signature(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
