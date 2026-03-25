// Package rest provides HTTP handlers and middleware for the REST API.
package rest

import (
	"bytes"
	"io"
	"net/http"

	pkgslack "github.com/Asheze1127/progress-checker/backend/pkg/slack"
	idempotencysvc "github.com/Asheze1127/progress-checker/backend/service/idempotency"
)

// headerSlackRetryNum is the Slack header indicating a retry.
const headerSlackRetryNum = "X-Slack-Retry-Num"

// slackIdempotencyKey extracts an idempotency key from a Slack request.
// Returns an empty string for first attempts (no retry header).
func slackIdempotencyKey(r *http.Request) string {
	retryNum := r.Header.Get(headerSlackRetryNum)
	if retryNum == "" {
		return ""
	}
	return r.URL.Path + ":retry:" + retryNum
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
			key := slackIdempotencyKey(r)
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
