package webhook

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Asheze1127/progress-checker/backend/application/service/question_sender"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// testQuestionRepository is a test double for QuestionRepository.
type testQuestionRepository struct {
	saveErr error
}

func (m *testQuestionRepository) Save(_ context.Context, _ *entities.Question) error {
	return m.saveErr
}

func (m *testQuestionRepository) GetByID(_ context.Context, _ entities.QuestionID) (*entities.Question, error) {
	return nil, nil
}

func (m *testQuestionRepository) GetByThreadTS(_ context.Context, _ entities.SlackChannelID, _ string) (*entities.Question, error) {
	return nil, nil
}

func (m *testQuestionRepository) GetAwaitingByChannelAndThread(_ context.Context, _ entities.SlackChannelID, _ string) (*entities.Question, error) {
	return nil, nil
}

func (m *testQuestionRepository) UpdateStatus(_ context.Context, _ entities.QuestionID, _ entities.QuestionStatus) error {
	return nil
}

func (m *testQuestionRepository) AssignMentor(_ context.Context, _ entities.QuestionID, _ entities.MentorID) error {
	return nil
}

// testMessageQueue is a test double for MessageQueue.
type testMessageQueue struct {
	sendErr error
}

func (m *testMessageQueue) Send(_ context.Context, _ string, _ []byte) error {
	return m.sendErr
}

func newTestHandler(queueErr error) *QuestionHandler {
	repo := &testQuestionRepository{}
	queue := &testMessageQueue{sendErr: queueErr}
	sender := questionsender.NewQuestionSender(queue)
	uc := usecase.NewHandleNewQuestionUseCase(repo, sender)
	return NewQuestionHandler(uc)
}

func newTestHandlerWithFailingRepo() *QuestionHandler {
	repo := &testQuestionRepository{saveErr: errors.New("database error")}
	queue := &testMessageQueue{}
	sender := questionsender.NewQuestionSender(queue)
	uc := usecase.NewHandleNewQuestionUseCase(repo, sender)
	return NewQuestionHandler(uc)
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

func init() {
	gin.SetMode(gin.TestMode)
}

func callHandler(handler *QuestionHandler, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/webhook/slack/questions", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler.HandleWebhook(c)
	return rec
}

func TestHandleWebhook_NewQuestionSuccess(t *testing.T) {
	handler := newTestHandler(nil)
	body := makeSlackForm("/question", "How do I deploy?", "U123", "C456", "")
	rec := callHandler(handler, body)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	respBody := rec.Body.String()
	if !strings.Contains(respBody, "Your question has been received") {
		t.Errorf("expected acknowledgment message, got %q", respBody)
	}
}

func TestHandleWebhook_UnsupportedCommand(t *testing.T) {
	handler := newTestHandler(nil)
	body := makeSlackForm("/progress", "some text", "U123", "C456", "")
	rec := callHandler(handler, body)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandleWebhook_FollowUpReturnsOK(t *testing.T) {
	handler := newTestHandler(nil)
	body := makeSlackForm("/question", "More info here", "U123", "C456", "1234567890.123456")
	rec := callHandler(handler, body)

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
	rec := callHandler(handler, body)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandleWebhook_MissingUserID(t *testing.T) {
	handler := newTestHandler(nil)
	body := makeSlackForm("/question", "Some question", "", "C456", "")
	rec := callHandler(handler, body)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandleWebhook_MissingChannelID(t *testing.T) {
	handler := newTestHandler(nil)
	body := makeSlackForm("/question", "Some question", "U123", "", "")
	rec := callHandler(handler, body)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandleWebhook_ServiceError(t *testing.T) {
	handler := newTestHandlerWithFailingRepo()
	body := makeSlackForm("/question", "How do I deploy?", "U123", "C456", "")
	rec := callHandler(handler, body)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestHandleWebhook_QueueError(t *testing.T) {
	handler := newTestHandler(errors.New("SQS error"))
	body := makeSlackForm("/question", "How do I deploy?", "U123", "C456", "")
	rec := callHandler(handler, body)

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
			name:     "multibyte text truncated at rune boundary",
			input:    strings.Repeat("あ", maxTitleLength+5),
			expected: strings.Repeat("あ", maxTitleLength) + "...",
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
