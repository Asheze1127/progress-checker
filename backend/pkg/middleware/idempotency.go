// Package middleware provides HTTP middleware components.
package middleware

import (
	"context"
	"log"
	"net/http"
	"time"
)

// defaultIdempotencyTTL is the default time-to-live for idempotency keys.
const defaultIdempotencyTTL = 1 * time.Hour

// headerSlackRetryNum is the Slack header indicating a retry.
const headerSlackRetryNum = "X-Slack-Retry-Num"

// IdempotencyStore defines the interface for checking and storing idempotency keys.
type IdempotencyStore interface {
	// Exists checks whether the given key has already been processed.
	Exists(ctx context.Context, key string) (bool, error)

	// Set stores the key with the specified TTL.
	Set(ctx context.Context, key string, ttl time.Duration) error
}

// IdempotencyKeyFunc extracts an idempotency key from the request.
// Return an empty string to skip idempotency checking for the request.
type IdempotencyKeyFunc func(r *http.Request) string

// DefaultIdempotencyKeyFunc extracts the idempotency key by combining
// X-Slack-Retry-Num with the request body event_id or trigger_id.
// For simplicity, it uses X-Slack-Retry-Num as the signal and the
// full request URI + retry number as the composite key.
func DefaultIdempotencyKeyFunc(r *http.Request) string {
	retryNum := r.Header.Get(headerSlackRetryNum)
	if retryNum == "" {
		// First attempt, no retry header present.
		// We still generate a key based on the request path to track it.
		return ""
	}
	// For retries, build a key that includes the retry info.
	return r.URL.Path + ":retry:" + retryNum
}

// Idempotency returns middleware that checks for duplicate requests using the provided store.
// If keyFunc is nil, DefaultIdempotencyKeyFunc is used.
func Idempotency(store IdempotencyStore, keyFunc IdempotencyKeyFunc) func(http.Handler) http.Handler {
	if keyFunc == nil {
		keyFunc = DefaultIdempotencyKeyFunc
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)
			if key == "" {
				// No idempotency key extracted; proceed without checking.
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()

			exists, err := store.Exists(ctx, key)
			if err != nil {
				// Log the error but continue processing to avoid blocking
				// legitimate requests due to store failures.
				log.Printf("idempotency store error on Exists: %v", err)
				next.ServeHTTP(w, r)
				return
			}

			if exists {
				// Duplicate request; return 200 OK immediately.
				w.WriteHeader(http.StatusOK)
				return
			}

			if err := store.Set(ctx, key, defaultIdempotencyTTL); err != nil {
				// Log the error but continue processing.
				log.Printf("idempotency store error on Set: %v", err)
			}

			next.ServeHTTP(w, r)
		})
	}
}
