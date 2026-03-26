package middleware

import (
	"bytes"
	"io"
	"net/http"

	pkgslack "github.com/Asheze1127/progress-checker/backend/pkg/slack"
	idempotencysvc "github.com/Asheze1127/progress-checker/backend/service/idempotency"
)

// slackDeduplicationKey extracts a deduplication key from a Slack request.
// Uses trigger_id from the form body for slash commands and interactions.
// Falls back to X-Slack-Request-Timestamp + URL path for Events API callbacks
// which don't include trigger_id.
func slackDeduplicationKey(r *http.Request) string {
	if err := r.ParseForm(); err != nil {
		return ""
	}

	triggerID := r.FormValue("trigger_id")
	if triggerID != "" {
		return "slack:" + triggerID
	}

	// Fallback for Events API callbacks: use request timestamp + path as key.
	// Slack retries send X-Slack-Retry-Num header, but the original request
	// also needs a key, so we use the timestamp which is unique per delivery.
	ts := r.Header.Get("X-Slack-Request-Timestamp")
	if ts != "" {
		return "slack:event:" + ts + ":" + r.URL.Path
	}

	return ""
}

// SlackVerification returns middleware that verifies Slack request signatures.
// Verified request bodies are restored so downstream handlers can read them.
func SlackVerification(verifier *pkgslack.Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := verifier.Verify(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
		})
	}
}

// SlackIdempotency returns middleware that checks for duplicate Slack requests
// using the idempotency service.
func SlackIdempotency(svc *idempotencysvc.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := slackDeduplicationKey(r)
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			if svc.IsDuplicate(r.Context(), key) {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SlackWebhookMiddleware chains SlackVerification and SlackIdempotency middleware
// in the correct order: signature verification first, then idempotency check.
func SlackWebhookMiddleware(
	verifier *pkgslack.Verifier,
	svc *idempotencysvc.Service,
) func(http.Handler) http.Handler {
	verifyMiddleware := SlackVerification(verifier)
	idempotencyMiddleware := SlackIdempotency(svc)

	return func(handler http.Handler) http.Handler {
		return verifyMiddleware(idempotencyMiddleware(handler))
	}
}
