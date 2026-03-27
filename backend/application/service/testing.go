package service

import (
	"github.com/Asheze1127/progress-checker/backend/application/port"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// NewProgressFormatterForTest creates a ProgressFormatter for testing without DI.
func NewProgressFormatterForTest() *ProgressFormatter {
	return &ProgressFormatter{}
}

// NewPasswordHasherForTest creates a PasswordHasher for testing without DI.
func NewPasswordHasherForTest() *PasswordHasher {
	return &PasswordHasher{}
}

// NewJWTServiceForTest creates a JWTService for testing without DI.
func NewJWTServiceForTest(secret string) *JWTService {
	return newJWTService(secret)
}

// NewSlackPosterForTest creates a SlackPoster for testing without DI.
func NewSlackPosterForTest(client SlackClient) *SlackPoster {
	return &SlackPoster{
		client:    client,
		formatter: &ProgressFormatter{},
	}
}

// NewQuestionSenderForTest creates a QuestionSender for testing without DI.
func NewQuestionSenderForTest(queue port.MessageQueue) *QuestionSender {
	return &QuestionSender{queue: queue}
}

// NewGitHubServiceForTest creates a GitHubService for testing without DI.
func NewGitHubServiceForTest(
	repo entities.GitHubRepoRepository,
	encryptor port.TokenEncryptor,
	ghClient port.GitHubIssueCreator,
	idProvider func() string,
) *GitHubService {
	return &GitHubService{
		repo:       repo,
		encryptor:  encryptor,
		ghClient:   ghClient,
		idProvider: idProvider,
	}
}
