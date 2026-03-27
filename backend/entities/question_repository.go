package entities

import "context"

// QuestionRepository defines the interface for persisting and querying questions.
type QuestionRepository interface {
  Save(ctx context.Context, question *Question) error
  GetByID(ctx context.Context, id QuestionID) (*Question, error)
  GetByThreadTS(ctx context.Context, channelID SlackChannelID, threadTS string) (*Question, error)
  GetAwaitingByChannelAndThread(ctx context.Context, channelID SlackChannelID, threadTS string) (*Question, error)
  UpdateStatus(ctx context.Context, id QuestionID, status QuestionStatus) error
  AssignMentor(ctx context.Context, questionID QuestionID, mentorUserID MentorID) error
}
