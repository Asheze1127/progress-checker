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
	now := time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)
	currentTimestamp := fmt.Sprintf("%d", now.Unix())

	tests := []struct {
		name           string
		signature      string
		timestamp      string
		body           string
		wantErr        error
		wantHTTPStatus int
	}{
		{
			name:           "valid signature passes verification",
			signature:      computeSignature(signingSecret, currentTimestamp, `{"event":"test"}`),
			timestamp:      currentTimestamp,
			body:           `{"event":"test"}`,
			wantErr:        nil,
			wantHTTPStatus: 0,
		},
		{
			name:           "missing signature header returns error",
			signature:      "",
			timestamp:      currentTimestamp,
			body:           `{"event":"test"}`,
			wantErr:        ErrMissingSignature,
			wantHTTPStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing timestamp header returns error",
			signature:      "v0=abc123",
			timestamp:      "",
			body:           `{"event":"test"}`,
			wantErr:        ErrMissingTimestamp,
			wantHTTPStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid timestamp format returns error",
			signature:      "v0=abc123",
			timestamp:      "not-a-number",
			body:           `{"event":"test"}`,
			wantErr:        ErrInvalidTimestamp,
			wantHTTPStatus: http.StatusUnauthorized,
		},
		{
			name:           "expired timestamp returns forbidden",
			signature:      "v0=abc123",
			timestamp:      fmt.Sprintf("%d", now.Add(-6*time.Minute).Unix()),
			body:           `{"event":"test"}`,
			wantErr:        ErrExpiredTimestamp,
			wantHTTPStatus: http.StatusForbidden,
		},
		{
			name:           "future timestamp beyond threshold returns forbidden",
			signature:      "v0=abc123",
			timestamp:      fmt.Sprintf("%d", now.Add(6*time.Minute).Unix()),
			body:           `{"event":"test"}`,
			wantErr:        ErrExpiredTimestamp,
			wantHTTPStatus: http.StatusForbidden,
		},
		{
			name:           "invalid signature returns unauthorized",
			signature:      "v0=invalidsignaturevalue",
			timestamp:      currentTimestamp,
			body:           `{"event":"test"}`,
			wantErr:        ErrInvalidSignature,
			wantHTTPStatus: http.StatusUnauthorized,
		},
		{
			name:           "wrong secret produces invalid signature",
			signature:      computeSignature("wrong-secret", currentTimestamp, `{"event":"test"}`),
			timestamp:      currentTimestamp,
			body:           `{"event":"test"}`,
			wantErr:        ErrInvalidSignature,
			wantHTTPStatus: http.StatusUnauthorized,
		},
		{
			name:           "tampered body produces invalid signature",
			signature:      computeSignature(signingSecret, currentTimestamp, `{"event":"test"}`),
			timestamp:      currentTimestamp,
			body:           `{"event":"tampered"}`,
			wantErr:        ErrInvalidSignature,
			wantHTTPStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier := NewVerifier(signingSecret)
			verifier.nowFunc = func() time.Time { return now }

			req := httptest.NewRequest(http.MethodPost, "/webhook/slack", strings.NewReader(tt.body))
			if tt.signature != "" {
				req.Header.Set(headerSignature, tt.signature)
			}
			if tt.timestamp != "" {
				req.Header.Set(headerTimestamp, tt.timestamp)
			}

			body, err := verifier.Verify(req)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				if got := HTTPStatusForError(err); got != tt.wantHTTPStatus {
					t.Fatalf("expected HTTP status %d, got %d", tt.wantHTTPStatus, got)
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
