// Package middleware provides HTTP middleware components.
package middleware

import (
	"net/http"

	idempotencysvc "github.com/Asheze1127/progress-checker/backend/service/idempotency"
)

// headerSlackRetryNum is the Slack header indicating a retry.
const headerSlackRetryNum = "X-Slack-Retry-Num"

// IdempotencyKeyFunc extracts an idempotency key from the request.
// Return an empty string to skip idempotency checking for the request.
type IdempotencyKeyFunc func(r *http.Request) string

// DefaultIdempotencyKeyFunc extracts the idempotency key by combining
// X-Slack-Retry-Num with the request URI.
// Returns an empty string for first attempts (no retry header).
func DefaultIdempotencyKeyFunc(r *http.Request) string {
	retryNum := r.Header.Get(headerSlackRetryNum)
	if retryNum == "" {
		return ""
	}
	return r.URL.Path + ":retry:" + retryNum
}

// Idempotency returns middleware that checks for duplicate requests using the idempotency service.
// If keyFunc is nil, DefaultIdempotencyKeyFunc is used.
func Idempotency(svc *idempotencysvc.Service, keyFunc IdempotencyKeyFunc) func(http.Handler) http.Handler {
	if keyFunc == nil {
		keyFunc = DefaultIdempotencyKeyFunc
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)
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
