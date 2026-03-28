// Package slack provides utilities for Slack webhook request verification.
package slack

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/slack-go/slack"
)

var (
	// ErrInvalidSignature indicates the computed HMAC does not match the provided signature.
	ErrInvalidSignature = errors.New("invalid request signature")
)

// Verifier validates Slack request signatures using the official slack-go library.
type Verifier struct {
	signingSecret string
}

// NewVerifier creates a new Verifier with the given Slack signing secret.
// Returns an error if the signing secret is empty.
func NewVerifier(signingSecret string) (*Verifier, error) {
	if len(signingSecret) == 0 {
		return nil, errors.New("slack signing secret must not be empty")
	}
	return &Verifier{
		signingSecret: signingSecret,
	}, nil
}

// Verify checks that the request has a valid Slack signature using slack-go's SecretsVerifier.
// It reads and returns the request body so callers can continue processing.
func (v *Verifier) Verify(r *http.Request) ([]byte, error) {
	sv, err := slack.NewSecretsVerifier(r.Header, v.signingSecret)
	if err != nil {
		return nil, ErrInvalidSignature
	}

	const maxSlackBodyBytes = 2 << 20 // 2 MB
	body, err := io.ReadAll(io.TeeReader(io.LimitReader(r.Body, maxSlackBodyBytes), &sv))
	if err != nil {
		return nil, fmt.Errorf("reading request body: %w", err)
	}

	if err := sv.Ensure(); err != nil {
		return nil, ErrInvalidSignature
	}

	return body, nil
}
