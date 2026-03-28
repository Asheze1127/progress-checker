package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func computeSignature(secret, timestamp, body string) string {
	baseString := fmt.Sprintf("v0:%s:%s", timestamp, body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(baseString))
	return "v0=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifier_Verify(t *testing.T) {
	const signingSecret = "test-signing-secret"
	now := time.Now()
	currentTimestamp := fmt.Sprintf("%d", now.Unix())

	tests := []struct {
		name      string
		signature string
		timestamp string
		body      string
		wantErr   bool
	}{
		{
			name:      "valid signature passes verification",
			signature: computeSignature(signingSecret, currentTimestamp, `{"event":"test"}`),
			timestamp: currentTimestamp,
			body:      `{"event":"test"}`,
			wantErr:   false,
		},
		{
			name:      "missing signature header returns error",
			signature: "",
			timestamp: currentTimestamp,
			body:      `{"event":"test"}`,
			wantErr:   true,
		},
		{
			name:      "missing timestamp header returns error",
			signature: "v0=abc123",
			timestamp: "",
			body:      `{"event":"test"}`,
			wantErr:   true,
		},
		{
			name:      "expired timestamp returns error",
			signature: "v0=abc123",
			timestamp: fmt.Sprintf("%d", now.Add(-6*time.Minute).Unix()),
			body:      `{"event":"test"}`,
			wantErr:   true,
		},
		{
			name:      "invalid signature returns error",
			signature: "v0=invalidsignaturevalue",
			timestamp: currentTimestamp,
			body:      `{"event":"test"}`,
			wantErr:   true,
		},
		{
			name:      "wrong secret produces invalid signature",
			signature: computeSignature("wrong-secret", currentTimestamp, `{"event":"test"}`),
			timestamp: currentTimestamp,
			body:      `{"event":"test"}`,
			wantErr:   true,
		},
		{
			name:      "tampered body produces invalid signature",
			signature: computeSignature(signingSecret, currentTimestamp, `{"event":"test"}`),
			timestamp: currentTimestamp,
			body:      `{"event":"tampered"}`,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier, err := NewVerifier(signingSecret)
			if err != nil {
				t.Fatalf("failed to create verifier: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/webhook/slack", strings.NewReader(tt.body))
			if tt.signature != "" {
				req.Header.Set("X-Slack-Signature", tt.signature)
			}
			if tt.timestamp != "" {
				req.Header.Set("X-Slack-Request-Timestamp", tt.timestamp)
			}

			body, err := verifier.Verify(req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(body) != tt.body {
				t.Fatalf("expected body %q, got %q", tt.body, string(body))
			}
		})
	}
}
