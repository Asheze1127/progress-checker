package questionsender

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Asheze1127/progress-checker/backend/application/service/message_queue"
)

const queueNameQuestionNew = "question-new"

type QuestionNewMessage struct {
	Type           string `json:"type"`
	QuestionID     string `json:"question_id"`
	ParticipantID  string `json:"participant_id"`
	Title          string `json:"title"`
	Text           string `json:"text"`
	SlackChannelID string `json:"slack_channel_id"`
	SlackThreadTS  string `json:"slack_thread_ts"`
}

type QuestionSender struct {
	queue messagequeue.MessageQueue
}

func NewQuestionSender(queue messagequeue.MessageQueue) *QuestionSender {
	return &QuestionSender{queue: queue}
}

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
