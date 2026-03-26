package service

import (
	"context"
	"encoding/json"
	"fmt"
)

// MessageQueue defines the interface for sending messages to a queue.
type MessageQueue interface {
	Send(ctx context.Context, queueName string, message []byte) error
}

const queueNameQuestionNew = "question:new"

// QuestionNewMessage represents the SQS message format for new questions.
type QuestionNewMessage struct {
	Type           string `json:"type"`
	QuestionID     string `json:"question_id"`
	ParticipantID  string `json:"participant_id"`
	Title          string `json:"title"`
	Text           string `json:"text"`
	SlackChannelID string `json:"slack_channel_id"`
	SlackThreadTS  string `json:"slack_thread_ts"`
}

// QuestionSender sends question-related messages to SQS.
type QuestionSender struct {
	queue MessageQueue
}

// NewQuestionSender creates a new QuestionSender with the given message queue.
func NewQuestionSender(queue MessageQueue) *QuestionSender {
	return &QuestionSender{queue: queue}
}

// SendNewQuestion sends a new question message to the SQS queue.
func (s *QuestionSender) SendNewQuestion(ctx context.Context, msg QuestionNewMessage) error {
	msg.Type = "question_new"

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := s.queue.Send(ctx, queueNameQuestionNew, msgBytes); err != nil {
		return fmt.Errorf("failed to enqueue message: %w", err)
	}

	return nil
}
