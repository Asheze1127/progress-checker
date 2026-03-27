package middleware

import (
	"bytes"
	"io"
	"net/http"

	pkgslack "github.com/Asheze1127/progress-checker/backend/pkg/slack"
)

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

// SlackRetryRejection returns middleware that rejects Slack retry requests.
// If the X-Slack-Retry-Num header is present, the request is a retry and
// we respond with 200 OK immediately to prevent duplicate processing.
func SlackRetryRejection() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Slack-Retry-Num") != "" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SlackWebhookMiddleware chains SlackVerification and SlackRetryRejection
// middleware in the correct order: signature verification first, then retry
// rejection.
func SlackWebhookMiddleware(
	verifier *pkgslack.Verifier,
) func(http.Handler) http.Handler {
	verifyMiddleware := SlackVerification(verifier)
	retryMiddleware := SlackRetryRejection()

	return func(handler http.Handler) http.Handler {
		return verifyMiddleware(retryMiddleware(handler))
	}
}
