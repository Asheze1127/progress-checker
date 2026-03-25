package usecase

import (
	"context"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// --- Test doubles ---

type stubQuestionRepo struct {
	question      *entities.Question
	findErr       error
	updateErr     error
	updatedID     entities.QuestionID
	updatedStatus entities.QuestionStatus
}

func (r *stubQuestionRepo) FindByID(_ context.Context, id entities.QuestionID) (*entities.Question, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	return r.question, nil
}

func (r *stubQuestionRepo) UpdateStatus(_ context.Context, id entities.QuestionID, status entities.QuestionStatus) error {
	r.updatedID = id
	r.updatedStatus = status
	return r.updateErr
}

type stubSlackNotifier struct {
	postErr  error
	posted   bool
	question *entities.Question
}

func (n *stubSlackNotifier) PostToMentorChannel(_ context.Context, q *entities.Question) error {
	n.posted = true
	n.question = q
	return n.postErr
}

func newTestQuestion(status entities.QuestionStatus) *entities.Question {
	return &entities.Question{
		ID:             "q-1",
		ParticipantID:  "p-1",
		Title:          "Test question",
		SlackChannelID: "C123",
		Status:         status,
		SlackThreadTS:  "1700000000.000100",
	}
}
