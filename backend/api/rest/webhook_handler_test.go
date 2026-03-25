package rest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Asheze1127/progress-checker/backend/application"
)

// mockProgressCommandHandler is a test double for ProgressCommandHandler.
type mockProgressCommandHandler struct {
	receivedInput application.ProgressCommandInput
	err           error
	called        bool
}

func (m *mockProgressCommandHandler) HandleProgressCommand(_ context.Context, input application.ProgressCommandInput) error {
	m.called = true
	m.receivedInput = input
	return m.err
}

func TestHandleWebhook(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		formValues       url.Values
		handlerErr       error
		wantStatus       int
		wantBodyContains string
		wantCalled       bool
		wantPhase        string
		wantSOS          bool
		wantComment      string
	}{
		{
			name:   "successful progress command",
			method: http.MethodPost,
			formValues: url.Values{
				"command":    {"/progress"},
				"user_id":    {"U12345"},
				"team_id":    {"T12345"},
				"channel_id": {"C12345"},
				"text":       {"phase:coding comment:Working on feature"},
			},
			wantStatus:       http.StatusOK,
			wantBodyContains: "進捗報告を受け付けました",
			wantCalled:       true,
			wantPhase:        "coding",
			wantComment:      "Working on feature",
		},
		{
			name:   "progress command with SOS",
			method: http.MethodPost,
			formValues: url.Values{
				"command":    {"/progress"},
				"user_id":    {"U12345"},
				"team_id":    {"T12345"},
				"channel_id": {"C12345"},
				"text":       {"phase:testing sos:true comment:Need help"},
			},
			wantStatus:       http.StatusOK,
			wantBodyContains: "進捗報告を受け付けました",
			wantCalled:       true,
			wantPhase:        "testing",
			wantSOS:          true,
			wantComment:      "Need help",
		},
		{
			name:   "handler returns error",
			method: http.MethodPost,
			formValues: url.Values{
				"command":    {"/progress"},
				"user_id":    {"U12345"},
				"team_id":    {"T12345"},
				"channel_id": {"C12345"},
				"text":       {"phase:coding comment:test"},
			},
			handlerErr:       errors.New("some error"),
			wantStatus:       http.StatusInternalServerError,
			wantBodyContains: "failed to process progress command",
			wantCalled:       true,
		},
		{
			name:       "method not allowed",
			method:     http.MethodGet,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "unsupported command",
			method: http.MethodPost,
			formValues: url.Values{
				"command": {"/unknown"},
			},
			wantStatus:       http.StatusBadRequest,
			wantBodyContains: "unsupported command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockProgressCommandHandler{err: tt.handlerErr}
			handler := NewWebhookHandler(mock)

			body := tt.formValues.Encode()
			req := httptest.NewRequest(tt.method, "/webhook/slack", strings.NewReader(body))
			if tt.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}

			rr := httptest.NewRecorder()
			handler.HandleWebhook(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			if tt.wantBodyContains != "" && !strings.Contains(rr.Body.String(), tt.wantBodyContains) {
				t.Fatalf("response body %q does not contain %q", rr.Body.String(), tt.wantBodyContains)
			}

			if mock.called != tt.wantCalled {
				t.Fatalf("expected handler called=%v, got %v", tt.wantCalled, mock.called)
			}

			if tt.wantCalled && tt.handlerErr == nil {
				if mock.receivedInput.Phase != tt.wantPhase {
					t.Fatalf("expected phase %q, got %q", tt.wantPhase, mock.receivedInput.Phase)
				}
				if mock.receivedInput.SOS != tt.wantSOS {
					t.Fatalf("expected SOS %v, got %v", tt.wantSOS, mock.receivedInput.SOS)
				}
				if mock.receivedInput.Comment != tt.wantComment {
					t.Fatalf("expected comment %q, got %q", tt.wantComment, mock.receivedInput.Comment)
				}
			}
		})
	}
}

func TestParseProgressText(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		wantPhase   string
		wantSOS     bool
		wantComment string
	}{
		{
			name:        "full format",
			text:        "phase:coding sos:true comment:Working on feature X",
			wantPhase:   "coding",
			wantSOS:     true,
			wantComment: "Working on feature X",
		},
		{
			name:        "phase and comment only",
			text:        "phase:design comment:Designing the architecture",
			wantPhase:   "design",
			wantSOS:     false,
			wantComment: "Designing the architecture",
		},
		{
			name:      "phase only",
			text:      "phase:idea",
			wantPhase: "idea",
			wantSOS:   false,
		},
		{
			name:        "sos false",
			text:        "phase:testing sos:false comment:Almost done",
			wantPhase:   "testing",
			wantSOS:     false,
			wantComment: "Almost done",
		},
		{
			name: "empty text",
			text: "",
		},
		{
			name:        "comment with empty value after prefix",
			text:        "phase:demo comment:",
			wantPhase:   "demo",
			wantComment: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phase, sos, comment := parseProgressText(tt.text)
			if phase != tt.wantPhase {
				t.Fatalf("expected phase %q, got %q", tt.wantPhase, phase)
			}
			if sos != tt.wantSOS {
				t.Fatalf("expected SOS %v, got %v", tt.wantSOS, sos)
			}
			if comment != tt.wantComment {
				t.Fatalf("expected comment %q, got %q", tt.wantComment, comment)
			}
		})
	}
}
