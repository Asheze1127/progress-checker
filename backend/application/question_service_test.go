package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// mockQuestionRepository is a test double for QuestionRepository.
type mockQuestionRepository struct {
	saveFunc         func(ctx context.Context, q *entities.Question) error
	findByThreadFunc func(ctx context.Context, channelID, threadTS string) (*entities.Question, error)
	savedQuestion    *entities.Question
}

func (m *mockQuestionRepository) Save(ctx context.Context, q *entities.Question) error {
	m.savedQuestion = q
	if m.saveFunc != nil {
		return m.saveFunc(ctx, q)
	}
	return nil
}

func (m *mockQuestionRepository) FindByThreadTS(ctx context.Context, channelID, threadTS string) (*entities.Question, error) {
	if m.findByThreadFunc != nil {
		return m.findByThreadFunc(ctx, channelID, threadTS)
	}
	return nil, nil
}

// mockMessageQueue is a test double for MessageQueue.
type mockMessageQueue struct {
	sendFunc  func(ctx context.Context, queueName string, message []byte) error
	sentQueue string
	sentMsg   []byte
}

func (m *mockMessageQueue) Send(ctx context.Context, queueName string, message []byte) error {
	m.sentQueue = queueName
	m.sentMsg = message
	if m.sendFunc != nil {
		return m.sendFunc(ctx, queueName, message)
	}
	return nil
}

// mockIDGenerator is a test double for IDGenerator.
type mockIDGenerator struct {
	ids   []string
	index int
}

func (m *mockIDGenerator) Generate() string {
	if m.index >= len(m.ids) {
		return "default-id"
	}
	id := m.ids[m.index]
	m.index++
	return id
}

func TestHandleNewQuestion_Success(t *testing.T) {
	repo := &mockQuestionRepository{}
	queue := &mockMessageQueue{}
	idGen := &mockIDGenerator{ids: []string{"question-123", "thread-456"}}
	service := NewQuestionService(repo, queue, idGen)

	input := NewQuestionInput{
		ParticipantID:  "user-1",
		Title:          "How to fix this error?",
		Text:           "I am getting a nil pointer exception when calling the API",
		SlackChannelID: "C123",
	}

	err := service.HandleNewQuestion(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify question was saved
	if repo.savedQuestion == nil {
		t.Fatal("expected question to be saved")
	}
	if string(repo.savedQuestion.ID) != "question-123" {
		t.Errorf("expected question ID %q, got %q", "question-123", repo.savedQuestion.ID)
	}
	if repo.savedQuestion.Status != entities.QuestionStatusOpen {
		t.Errorf("expected status %q, got %q", entities.QuestionStatusOpen, repo.savedQuestion.Status)
	}
	if string(repo.savedQuestion.ParticipantID) != "user-1" {
		t.Errorf("expected participant ID %q, got %q", "user-1", repo.savedQuestion.ParticipantID)
	}

	// Verify message was enqueued
	if queue.sentQueue != "question:new" {
		t.Errorf("expected queue %q, got %q", "question:new", queue.sentQueue)
	}

	var msg questionNewMessage
	if err := json.Unmarshal(queue.sentMsg, &msg); err != nil {
		t.Fatalf("failed to unmarshal sent message: %v", err)
	}
	if msg.Type != "question_new" {
		t.Errorf("expected message type %q, got %q", "question_new", msg.Type)
	}
	if msg.QuestionID != "question-123" {
		t.Errorf("expected question_id %q, got %q", "question-123", msg.QuestionID)
	}
	if msg.ParticipantID != "user-1" {
		t.Errorf("expected participant_id %q, got %q", "user-1", msg.ParticipantID)
	}
	if msg.Title != "How to fix this error?" {
		t.Errorf("expected title %q, got %q", "How to fix this error?", msg.Title)
	}
	if msg.Text != "I am getting a nil pointer exception when calling the API" {
		t.Errorf("expected text to match input")
	}
	if msg.SlackChannelID != "C123" {
		t.Errorf("expected slack_channel_id %q, got %q", "C123", msg.SlackChannelID)
	}
	if msg.SlackThreadTS != "thread-456" {
		t.Errorf("expected slack_thread_ts %q, got %q", "thread-456", msg.SlackThreadTS)
	}
}

func TestHandleNewQuestion_ValidationError(t *testing.T) {
	repo := &mockQuestionRepository{}
	queue := &mockMessageQueue{}
	idGen := &mockIDGenerator{ids: []string{"question-123", "thread-456"}}
	service := NewQuestionService(repo, queue, idGen)

	// Empty title should cause validation error
	input := NewQuestionInput{
		ParticipantID:  "user-1",
		Title:          "",
		Text:           "some text",
		SlackChannelID: "C123",
	}

	err := service.HandleNewQuestion(context.Background(), input)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	if repo.savedQuestion != nil {
		t.Error("expected no question to be saved on validation error")
	}
}

func TestHandleNewQuestion_SaveError(t *testing.T) {
	saveErr := errors.New("database connection failed")
	repo := &mockQuestionRepository{
		saveFunc: func(ctx context.Context, q *entities.Question) error {
			return saveErr
		},
	}
	queue := &mockMessageQueue{}
	idGen := &mockIDGenerator{ids: []string{"question-123", "thread-456"}}
	service := NewQuestionService(repo, queue, idGen)

	input := NewQuestionInput{
		ParticipantID:  "user-1",
		Title:          "How to fix this error?",
		Text:           "Full question text",
		SlackChannelID: "C123",
	}

	err := service.HandleNewQuestion(context.Background(), input)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, saveErr) {
		t.Errorf("expected wrapped save error, got %v", err)
	}

	// Verify message was NOT enqueued
	if queue.sentMsg != nil {
		t.Error("expected no message to be enqueued on save error")
	}
}

func TestHandleNewQuestion_QueueError(t *testing.T) {
	queueErr := errors.New("SQS unavailable")
	repo := &mockQuestionRepository{}
	queue := &mockMessageQueue{
		sendFunc: func(ctx context.Context, queueName string, message []byte) error {
			return queueErr
		},
	}
	idGen := &mockIDGenerator{ids: []string{"question-123", "thread-456"}}
	service := NewQuestionService(repo, queue, idGen)

	input := NewQuestionInput{
		ParticipantID:  "user-1",
		Title:          "How to fix this error?",
		Text:           "Full question text",
		SlackChannelID: "C123",
	}

	err := service.HandleNewQuestion(context.Background(), input)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, queueErr) {
		t.Errorf("expected wrapped queue error, got %v", err)
	}
}
