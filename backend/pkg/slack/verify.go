// Package slack provides utilities for Slack webhook request verification.
package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"
)

const (
	// headerSignature is the Slack request signature header.
	headerSignature = "X-Slack-Signature"

	// headerTimestamp is the Slack request timestamp header.
	headerTimestamp = "X-Slack-Request-Timestamp"

	// signatureVersion is the version prefix used by Slack for HMAC signatures.
	signatureVersion = "v0"

	// maxTimestampAge is the maximum allowed age for a request timestamp.
	// Requests older than this are rejected to prevent replay attacks.
	maxTimestampAge = 5 * time.Minute
)

var (
	// ErrMissingSignature indicates the X-Slack-Signature header is absent.
	ErrMissingSignature = errors.New("missing X-Slack-Signature header")

	// ErrMissingTimestamp indicates the X-Slack-Request-Timestamp header is absent.
	ErrMissingTimestamp = errors.New("missing X-Slack-Request-Timestamp header")

	// ErrInvalidTimestamp indicates the timestamp header is not a valid Unix timestamp.
	ErrInvalidTimestamp = errors.New("invalid X-Slack-Request-Timestamp value")

	// ErrExpiredTimestamp indicates the request timestamp is too old.
	ErrExpiredTimestamp = errors.New("request timestamp is too old")

	// ErrInvalidSignature indicates the computed HMAC does not match the provided signature.
	ErrInvalidSignature = errors.New("invalid request signature")
)

// Verifier validates Slack request signatures using HMAC-SHA256.
type Verifier struct {
	signingSecret string
	nowFunc       func() time.Time // for testing
}

// NewVerifier creates a new Verifier with the given Slack signing secret.
func NewVerifier(signingSecret string) *Verifier {
	return &Verifier{
		signingSecret: signingSecret,
		nowFunc:       time.Now,
	}
}

// Verify checks that the request has a valid Slack signature.
// It reads and returns the request body so callers can continue processing.
// Returns the body bytes on success, or an error with an appropriate HTTP status.
func (v *Verifier) Verify(r *http.Request) ([]byte, error) {
	signature := r.Header.Get(headerSignature)
	if signature == "" {
		return nil, ErrMissingSignature
	}

	timestampStr := r.Header.Get(headerTimestamp)
	if timestampStr == "" {
		return nil, ErrMissingTimestamp
	}

	timestampInt, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return nil, ErrInvalidTimestamp
	}

	requestTime := time.Unix(timestampInt, 0)
	age := v.nowFunc().Sub(requestTime)
	if math.Abs(age.Seconds()) > maxTimestampAge.Seconds() {
		return nil, ErrExpiredTimestamp
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("reading request body: %w", err)
	}

	baseString := fmt.Sprintf("%s:%s:%s", signatureVersion, timestampStr, string(body))

	mac := hmac.New(sha256.New, []byte(v.signingSecret))
	mac.Write([]byte(baseString))
	expectedSignature := signatureVersion + "=" + hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expectedSignature), []byte(signature)) {
		return nil, ErrInvalidSignature
	}

	return body, nil
}

// HTTPStatusForError returns the appropriate HTTP status code for a verification error.
func HTTPStatusForError(err error) int {
	switch {
	case errors.Is(err, ErrExpiredTimestamp):
		return http.StatusForbidden
	case errors.Is(err, ErrMissingSignature),
		errors.Is(err, ErrMissingTimestamp),
		errors.Is(err, ErrInvalidTimestamp),
		errors.Is(err, ErrInvalidSignature):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
