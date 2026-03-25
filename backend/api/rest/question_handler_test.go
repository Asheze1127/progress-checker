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
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// testQuestionRepository is a test double for QuestionRepository.
type testQuestionRepository struct {
	savedQuestion *entities.Question
}

func (m *testQuestionRepository) Save(ctx context.Context, q *entities.Question) error {
	m.savedQuestion = q
	return nil
}

func (m *testQuestionRepository) FindByThreadTS(ctx context.Context, channelID, threadTS string) (*entities.Question, error) {
	return nil, nil
}

// testMessageQueue is a test double for MessageQueue.
type testMessageQueue struct {
	sendErr error
}

func (m *testMessageQueue) Send(ctx context.Context, queueName string, message []byte) error {
	return m.sendErr
}

// testIDGenerator is a test double for IDGenerator.
type testIDGenerator struct {
	ids   []string
	index int
}

func (m *testIDGenerator) Generate() string {
	if m.index >= len(m.ids) {
		return "default-id"
	}
	id := m.ids[m.index]
	m.index++
	return id
}

func newTestHandler(queueErr error) *QuestionHandler {
	repo := &testQuestionRepository{}
	queue := &testMessageQueue{sendErr: queueErr}
	idGen := &testIDGenerator{ids: []string{"q-1", "t-1"}}
	service := application.NewQuestionService(repo, queue, idGen)
	return NewQuestionHandler(service)
}

func newTestHandlerWithFailingRepo() *QuestionHandler {
	repo := &failingQuestionRepository{}
	queue := &testMessageQueue{}
	idGen := &testIDGenerator{ids: []string{"q-1", "t-1"}}
	service := application.NewQuestionService(repo, queue, idGen)
	return NewQuestionHandler(service)
}

// failingQuestionRepository always returns an error on Save.
type failingQuestionRepository struct{}

func (m *failingQuestionRepository) Save(ctx context.Context, q *entities.Question) error {
	return errors.New("database error")
}

func (m *failingQuestionRepository) FindByThreadTS(ctx context.Context, channelID, threadTS string) (*entities.Question, error) {
	return nil, nil
}

func makeSlackForm(command, text, userID, channelID, threadTS string) string {
	v := url.Values{}
	v.Set("command", command)
	v.Set("text", text)
	v.Set("user_id", userID)
	v.Set("channel_id", channelID)
	if threadTS != "" {
		v.Set("thread_ts", threadTS)
	}
	return v.Encode()
}

func TestHandleWebhook_NewQuestionSuccess(t *testing.T) {
	handler := newTestHandler(nil)
	body := makeSlackForm("/question", "How do I deploy?", "U123", "C456", "")
	req := httptest.NewRequest(http.MethodPost, "/webhook/slack", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	respBody := rec.Body.String()
	if !strings.Contains(respBody, "Your question has been received") {
		t.Errorf("expected acknowledgment message, got %q", respBody)
	}
}

func TestHandleWebhook_MethodNotAllowed(t *testing.T) {
	handler := newTestHandler(nil)
	req := httptest.NewRequest(http.MethodGet, "/webhook/slack", nil)
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestHandleWebhook_UnsupportedCommand(t *testing.T) {
	handler := newTestHandler(nil)
	body := makeSlackForm("/progress", "some text", "U123", "C456", "")
	req := httptest.NewRequest(http.MethodPost, "/webhook/slack", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandleWebhook_FollowUpReturnsOK(t *testing.T) {
	handler := newTestHandler(nil)
	body := makeSlackForm("/question", "More info here", "U123", "C456", "1234567890.123456")
	req := httptest.NewRequest(http.MethodPost, "/webhook/slack", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	respBody := rec.Body.String()
	if !strings.Contains(respBody, "not yet supported") {
		t.Errorf("expected follow-up not supported message, got %q", respBody)
	}
}

func TestHandleWebhook_EmptyText(t *testing.T) {
	handler := newTestHandler(nil)
	body := makeSlackForm("/question", "", "U123", "C456", "")
	req := httptest.NewRequest(http.MethodPost, "/webhook/slack", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandleWebhook_MissingUserID(t *testing.T) {
	handler := newTestHandler(nil)
	body := makeSlackForm("/question", "Some question", "", "C456", "")
	req := httptest.NewRequest(http.MethodPost, "/webhook/slack", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandleWebhook_MissingChannelID(t *testing.T) {
	handler := newTestHandler(nil)
	body := makeSlackForm("/question", "Some question", "U123", "", "")
	req := httptest.NewRequest(http.MethodPost, "/webhook/slack", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandleWebhook_ServiceError(t *testing.T) {
	handler := newTestHandlerWithFailingRepo()
	body := makeSlackForm("/question", "How do I deploy?", "U123", "C456", "")
	req := httptest.NewRequest(http.MethodPost, "/webhook/slack", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestHandleWebhook_QueueError(t *testing.T) {
	handler := newTestHandler(errors.New("SQS error"))
	body := makeSlackForm("/question", "How do I deploy?", "U123", "C456", "")
	req := httptest.NewRequest(http.MethodPost, "/webhook/slack", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.HandleWebhook(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestTruncateTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short text unchanged",
			input:    "Short question",
			expected: "Short question",
		},
		{
			name:     "exactly max length unchanged",
			input:    strings.Repeat("a", maxTitleLength),
			expected: strings.Repeat("a", maxTitleLength),
		},
		{
			name:     "long text truncated with ellipsis",
			input:    strings.Repeat("a", maxTitleLength+10),
			expected: strings.Repeat("a", maxTitleLength) + "...",
		},
		{
			name:     "empty text unchanged",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateTitle(tt.input)
			if result != tt.expected {
				t.Errorf("truncateTitle(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
