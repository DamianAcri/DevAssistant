package test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DamianAcri/DevAssistant/internal/api"
	"github.com/DamianAcri/DevAssistant/internal/config"
)

// TestWebhookSignatureValid verifies that a correctly signed request
// passes middleware verification and reaches the handler.
func TestWebhookSignatureValid(t *testing.T) {
	secret := "test-secret"
	body := []byte(`{"action":"opened","pull_request":{"title":"Add feature"},"repository":{"full_name":"example/repo"}}`)

	// Build the full router so we test middleware + handler integration.
	router := api.NewRouter(config.Config{GitHubWebhookSecret: secret})
	req := httptest.NewRequest(http.MethodPost, "/webhook/github", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "pull_request")
	// Signature generated with the same secret configured in the app.
	req.Header.Set("X-Hub-Signature-256", "sha256="+computeHMACSHA256Hex(body, secret))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

// TestWebhookSignatureInvalid verifies that a request signed with the wrong
// secret is rejected by the middleware.
func TestWebhookSignatureInvalid(t *testing.T) {
	secret := "test-secret"
	body := []byte(`{"action":"opened","pull_request":{"title":"Add feature"},"repository":{"full_name":"example/repo"}}`)

	router := api.NewRouter(config.Config{GitHubWebhookSecret: secret})
	req := httptest.NewRequest(http.MethodPost, "/webhook/github", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "pull_request")
	// Signature intentionally generated with a different secret.
	req.Header.Set("X-Hub-Signature-256", "sha256="+computeHMACSHA256Hex(body, "wrong-secret"))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d. body=%s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}
}

// TestWebhookUnknownEvent verifies that unsupported event types are ignored
// with HTTP 200, as required to avoid unnecessary GitHub retries.
func TestWebhookUnknownEvent(t *testing.T) {
	secret := "test-secret"
	body := []byte(`{"hello":"world"}`)

	router := api.NewRouter(config.Config{GitHubWebhookSecret: secret})
	req := httptest.NewRequest(http.MethodPost, "/webhook/github", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", "sha256="+computeHMACSHA256Hex(body, secret))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

// computeHMACSHA256Hex returns the GitHub-style SHA256 HMAC signature payload.
func computeHMACSHA256Hex(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
