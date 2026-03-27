package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	pkgslack "github.com/Asheze1127/progress-checker/backend/pkg/slack"
)

// SlackVerification returns Gin middleware that verifies Slack request signatures.
// Verified request bodies are restored so downstream handlers can read them.
func SlackVerification(verifier *pkgslack.Verifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := verifier.Verify(c.Request)
		if err != nil {
			slog.Debug("SlackVerification: signature verification failed", slog.String("path", c.Request.URL.Path), slog.String("error", err.Error()))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		slog.Debug("slack request received",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("content_type", c.GetHeader("Content-Type")),
			slog.String("body", string(body)),
		)

		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		c.Next()
	}
}

// SlackRetryRejection returns Gin middleware that rejects Slack retry requests.
// If the X-Slack-Retry-Num header is present, the request is a retry and
// we respond with 200 OK immediately to prevent duplicate processing.
func SlackRetryRejection() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Slack-Retry-Num") != "" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	}
}
