package postgres

import (
	"context"
	"database/sql"

	db "github.com/Asheze1127/progress-checker/backend/database/postgres/generated"
	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/google/uuid"
)

var _ entities.QuestionRepository = (*QuestionRepository)(nil)

// QuestionRepository persists and queries questions in PostgreSQL.
type QuestionRepository struct {
	queries *db.Queries
}

// NewQuestionRepository creates a new QuestionRepository backed by the given database connection.
func NewQuestionRepository(database *sql.DB) *QuestionRepository {
	return &QuestionRepository{queries: db.New(database)}
}

func (r *QuestionRepository) Save(ctx context.Context, question *entities.Question) error {
	if err := question.Validate(); err != nil {
		return err
	}
	qID, err := uuid.Parse(string(question.ID))
	if err != nil {
		return err
	}
	pID, err := uuid.Parse(string(question.ParticipantID))
	if err != nil {
		return err
	}
	_, err = r.queries.InsertQuestion(ctx, db.InsertQuestionParams{
		ID:             qID,
		ParticipantID:  pID,
		Title:          question.Title,
		SlackChannelID: string(question.SlackChannelID),
		Status:         string(question.Status),
		SlackThreadTs:  question.SlackThreadTS,
	})
	return err
}

func (r *QuestionRepository) GetByID(ctx context.Context, id entities.QuestionID) (*entities.Question, error) {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return nil, err
	}
	row, err := r.queries.GetQuestionByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return toQuestionEntity(row), nil
}

func (r *QuestionRepository) GetByThreadTS(ctx context.Context, channelID entities.SlackChannelID, threadTS string) (*entities.Question, error) {
	row, err := r.queries.GetQuestionByThreadTS(ctx, db.GetQuestionByThreadTSParams{
		SlackChannelID: string(channelID),
		SlackThreadTs:  threadTS,
	})
	if err != nil {
		return nil, err
	}
	return toQuestionEntity(row), nil
}

func (r *QuestionRepository) GetAwaitingByChannelAndThread(ctx context.Context, channelID entities.SlackChannelID, threadTS string) (*entities.Question, error) {
	row, err := r.queries.GetAwaitingQuestionByChannelAndThread(ctx, db.GetAwaitingQuestionByChannelAndThreadParams{
		SlackChannelID: string(channelID),
		SlackThreadTs:  threadTS,
	})
	if err != nil {
		return nil, err
	}
	return toQuestionEntity(row), nil
}

func (r *QuestionRepository) UpdateStatus(ctx context.Context, id entities.QuestionID, status entities.QuestionStatus) error {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return err
	}
	return r.queries.UpdateQuestionStatus(ctx, db.UpdateQuestionStatusParams{
		Status: string(status),
		ID:     uid,
	})
}

func (r *QuestionRepository) AssignMentor(ctx context.Context, questionID entities.QuestionID, mentorUserID entities.MentorID) error {
	qID, err := uuid.Parse(string(questionID))
	if err != nil {
		return err
	}
	mID, err := uuid.Parse(string(mentorUserID))
	if err != nil {
		return err
	}
	return r.queries.InsertQuestionMentorAssignment(ctx, db.InsertQuestionMentorAssignmentParams{
		QuestionID:   qID,
		MentorUserID: mID,
	})
}

func toQuestionEntity(row db.Questions) *entities.Question {
	return &entities.Question{
		ID:             entities.QuestionID(row.ID.String()),
		ParticipantID:  entities.ParticipantID(row.ParticipantID.String()),
		Title:          row.Title,
		SlackChannelID: entities.SlackChannelID(row.SlackChannelID),
		Status:         entities.QuestionStatus(row.Status),
		SlackThreadTS:  row.SlackThreadTs,
	}
}
