// Package rest provides HTTP handlers and middleware for the REST API.
package rest

import (
	"bytes"
	"io"
	"net/http"

	"github.com/Asheze1127/progress-checker/backend/api/middleware"
	pkgslack "github.com/Asheze1127/progress-checker/backend/pkg/slack"
	idempotencysvc "github.com/Asheze1127/progress-checker/backend/service/idempotency"
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

			// Restore the body so downstream handlers can read them.
			r.Body = io.NopCloser(bytes.NewReader(body))

			next.ServeHTTP(w, r)
		})
	}
}

// SlackWebhookMiddleware chains SlackVerification and Idempotency middleware
// in the correct order: signature verification first, then idempotency check.
func SlackWebhookMiddleware(
	verifier *pkgslack.Verifier,
	svc *idempotencysvc.Service,
	keyFunc middleware.IdempotencyKeyFunc,
) func(http.Handler) http.Handler {
	verifyMiddleware := SlackVerification(verifier)
	idempotencyMiddleware := middleware.Idempotency(svc, keyFunc)

	return func(handler http.Handler) http.Handler {
		return verifyMiddleware(idempotencyMiddleware(handler))
	}
}
