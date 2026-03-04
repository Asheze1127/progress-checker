package slackapp

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/slack-go/slack"
)

func readVerifiedBody(r *http.Request, signingSecret string) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	sv, err := slack.NewSecretsVerifier(r.Header, signingSecret)
	if err != nil {
		return nil, err
	}

	if _, err := sv.Write(body); err != nil {
		return nil, err
	}
	if err := sv.Ensure(); err != nil {
		return nil, errors.New("signature verification failed")
	}

	return body, nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
