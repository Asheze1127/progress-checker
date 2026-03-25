package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// MessageQueue defines the interface for sending messages to a queue.
type MessageQueue interface {
	Send(ctx context.Context, queueName string, message []byte) error
}

// IDGenerator defines the interface for generating unique identifiers.
type IDGenerator interface {
	Generate() string
}

// QuestionService handles application-level logic for question operations.
type QuestionService struct {
	repo        entities.QuestionRepository
	queue       MessageQueue
	idGenerator IDGenerator
}

// NewQuestionService creates a new QuestionService with the given dependencies.
func NewQuestionService(repo entities.QuestionRepository, queue MessageQueue, idGenerator IDGenerator) *QuestionService {
	return &QuestionService{
		repo:        repo,
		queue:       queue,
		idGenerator: idGenerator,
	}
}

// NewQuestionInput contains the input data for creating a new question.
type NewQuestionInput struct {
	ParticipantID string
	Title         string
	Text          string
	SlackChannelID string
}

const queueNameQuestionNew = "question:new"

// questionNewMessage represents the SQS message format for new questions.
type questionNewMessage struct {
	Type           string `json:"type"`
	QuestionID     string `json:"question_id"`
	ParticipantID  string `json:"participant_id"`
	Title          string `json:"title"`
	Text           string `json:"text"`
	SlackChannelID string `json:"slack_channel_id"`
	SlackThreadTS  string `json:"slack_thread_ts"`
}

// HandleNewQuestion creates a new question record and enqueues it for processing.
func (s *QuestionService) HandleNewQuestion(ctx context.Context, input NewQuestionInput) error {
	questionID := s.idGenerator.Generate()
	threadTS := s.idGenerator.Generate()

	question := &entities.Question{
		ID:             entities.QuestionID(questionID),
		ParticipantID:  entities.ParticipantID(input.ParticipantID),
		Title:          input.Title,
		SlackChannelID: entities.SlackChannelID(input.SlackChannelID),
		Status:         entities.QuestionStatusOpen,
		SlackThreadTS:  threadTS,
	}

	if err := s.repo.Save(ctx, question); err != nil {
		return fmt.Errorf("failed to save question: %w", err)
	}

	msg := questionNewMessage{
		Type:           "question_new",
		QuestionID:     questionID,
		ParticipantID:  input.ParticipantID,
		Title:          input.Title,
		Text:           input.Text,
		SlackChannelID: input.SlackChannelID,
		SlackThreadTS:  threadTS,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := s.queue.Send(ctx, queueNameQuestionNew, msgBytes); err != nil {
		return fmt.Errorf("failed to enqueue message: %w", err)
	}

	return nil
}
