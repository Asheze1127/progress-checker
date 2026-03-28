package serve

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	_ "github.com/lib/pq"
	"github.com/samber/do/v2"

	"github.com/Asheze1127/progress-checker/backend/api/rest"
	"github.com/Asheze1127/progress-checker/backend/api/rest/handlers"
	"github.com/Asheze1127/progress-checker/backend/api/webhook"
	githubsvc "github.com/Asheze1127/progress-checker/backend/application/service/github"
	"github.com/Asheze1127/progress-checker/backend/application/service/jwt"
	"github.com/Asheze1127/progress-checker/backend/application/service/progress_formatter"
	"github.com/Asheze1127/progress-checker/backend/application/service/question_sender"
	"github.com/Asheze1127/progress-checker/backend/application/service/slack_poster"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/infrastructure/encryption"
	"github.com/Asheze1127/progress-checker/backend/infrastructure/githubclient"
	"github.com/Asheze1127/progress-checker/backend/infrastructure/postgres"
	slackinfra "github.com/Asheze1127/progress-checker/backend/infrastructure/slack"
	sqsinfra "github.com/Asheze1127/progress-checker/backend/infrastructure/sqs"
	pkgslack "github.com/Asheze1127/progress-checker/backend/pkg/slack"
	"github.com/Asheze1127/progress-checker/backend/util"
	"github.com/google/uuid"
)

// wireRouter builds all dependencies using samber/do DI container
// and returns the configured HTTP router.
func wireRouter(cfg *util.Config) (http.Handler, error) {
	injector := do.New()

	// --- Infrastructure ---
	do.Provide(injector, func(i do.Injector) (*sql.DB, error) {
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			url.PathEscape(cfg.DatabaseUser),
			url.PathEscape(cfg.DatabasePass),
			url.PathEscape(cfg.DatabaseHost),
			url.PathEscape(cfg.DatabasePort),
			url.PathEscape(cfg.DatabaseName),
			url.QueryEscape(cfg.DatabaseSSLMode),
		)
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
		if err := db.Ping(); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
		return db, nil
	})

	do.Provide(injector, func(i do.Injector) (*slackinfra.Client, error) {
		return slackinfra.NewClient(cfg.SlackBotToken), nil
	})

	do.Provide(injector, func(i do.Injector) (*encryption.AESEncryptor, error) {
		return encryption.NewAESEncryptor([]byte(cfg.EncryptionKey))
	})

	do.Provide(injector, func(i do.Injector) (*githubclient.Client, error) {
		return githubclient.NewClient(githubclient.Config{
			Token:             cfg.GitHubToken,
			BaseURL:           cfg.GitHubAPIBaseURL,
			AppIssuer:         cfg.GitHubAppID,
			AppInstallationID: cfg.GitHubAppInstallationID,
			AppPrivateKeyPEM:  cfg.GitHubAppPrivateKeyPEM,
		})
	})

	do.Provide(injector, func(i do.Injector) (*sqsinfra.Client, error) {
		awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion(cfg.AWSRegion),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}
		return sqsinfra.NewClient(sqs.NewFromConfig(awsCfg)), nil
	})

	do.Provide(injector, func(i do.Injector) (*slackinfra.MentorNotifier, error) {
		client := do.MustInvoke[*slackinfra.Client](i)
		return slackinfra.NewMentorNotifier(client.API(), cfg.SlackMentorChannelID), nil
	})

	// --- Repositories ---
	do.Provide(injector, func(i do.Injector) (*postgres.ProgressRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewProgressRepository(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*postgres.ProgressQueryRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewProgressQueryRepository(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*postgres.QuestionRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewQuestionRepository(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*postgres.UserRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewUserRepository(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*postgres.GitHubRepoRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewGitHubRepoRepository(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*postgres.StaffRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewStaffRepository(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*postgres.SetupTokenRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewSetupTokenRepository(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*postgres.ParticipantRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewParticipantRepository(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*postgres.MentorRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewMentorRepository(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*postgres.TeamRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewTeamRepository(db), nil
	})

	// --- Services ---
	do.Provide(injector, func(i do.Injector) (*progressformatter.ProgressFormatter, error) {
		return progressformatter.NewProgressFormatter(), nil
	})

	do.Provide(injector, func(i do.Injector) (*slackposter.SlackPoster, error) {
		client := do.MustInvoke[*slackinfra.Client](i)
		formatter := do.MustInvoke[*progressformatter.ProgressFormatter](i)
		return slackposter.NewSlackPoster(client, formatter), nil
	})

	do.Provide(injector, func(i do.Injector) (*jwt.JWTService, error) {
		return jwt.NewJWTService(cfg.JWTSecret)
	})

	do.Provide(injector, func(i do.Injector) (*util.PasswordHasher, error) {
		return util.NewPasswordHasher(), nil
	})

	do.Provide(injector, func(i do.Injector) (*pkgslack.Verifier, error) {
		return pkgslack.NewVerifier(cfg.SlackSigningSecret), nil
	})

	do.Provide(injector, func(i do.Injector) (*githubsvc.GitHubService, error) {
		ghRepoRepo := do.MustInvoke[*postgres.GitHubRepoRepository](i)
		encryptor := do.MustInvoke[*encryption.AESEncryptor](i)
		ghClient := do.MustInvoke[*githubclient.Client](i) // implements githubissuecreator.GitHubIssueCreator
		return githubsvc.NewGitHubService(ghRepoRepo, encryptor, ghClient, func() string {
			return uuid.New().String()
		}), nil
	})

	// --- Use Cases ---
	do.Provide(injector, func(i do.Injector) (*usecase.HandleProgressUseCase, error) {
		progressRepo := do.MustInvoke[*postgres.ProgressRepository](i)
		userRepo := do.MustInvoke[*postgres.UserRepository](i)
		poster := do.MustInvoke[*slackposter.SlackPoster](i)
		return usecase.NewHandleProgressUseCase(progressRepo, userRepo, poster), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.ListProgressUseCase, error) {
		progressQueryRepo := do.MustInvoke[*postgres.ProgressQueryRepository](i)
		mentorRepo := do.MustInvoke[*postgres.MentorRepository](i)
		return usecase.NewListProgressUseCase(progressQueryRepo, mentorRepo), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.LoginUseCase, error) {
		userRepo := do.MustInvoke[*postgres.UserRepository](i)
		jwtService := do.MustInvoke[*jwt.JWTService](i)
		hasher := do.MustInvoke[*util.PasswordHasher](i)
		return usecase.NewLoginUseCase(userRepo, jwtService, hasher), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.ResolveQuestionUseCase, error) {
		questionRepo := do.MustInvoke[*postgres.QuestionRepository](i)
		return usecase.NewResolveQuestionUseCase(questionRepo), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.ContinueQuestionUseCase, error) {
		questionRepo := do.MustInvoke[*postgres.QuestionRepository](i)
		return usecase.NewContinueQuestionUseCase(questionRepo), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.EscalateQuestionUseCase, error) {
		questionRepo := do.MustInvoke[*postgres.QuestionRepository](i)
		notifier := do.MustInvoke[*slackinfra.MentorNotifier](i)
		return usecase.NewEscalateQuestionUseCase(questionRepo, notifier), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.HandleNewQuestionUseCase, error) {
		questionRepo := do.MustInvoke[*postgres.QuestionRepository](i)
		sqsClient := do.MustInvoke[*sqsinfra.Client](i)
		sender := questionsender.NewQuestionSender(sqsClient)
		return usecase.NewHandleNewQuestionUseCase(questionRepo, sender), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.TriggerIssueCreationUseCase, error) {
		slackClient := do.MustInvoke[*slackinfra.Client](i)
		sqsClient := do.MustInvoke[*sqsinfra.Client](i)
		return usecase.NewTriggerIssueCreationUseCase(slackClient, sqsClient), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.StaffLoginUseCase, error) {
		staffRepo := do.MustInvoke[*postgres.StaffRepository](i)
		jwtService := do.MustInvoke[*jwt.JWTService](i)
		hasher := do.MustInvoke[*util.PasswordHasher](i)
		return usecase.NewStaffLoginUseCase(staffRepo, jwtService, hasher), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.CreateTeamUseCase, error) {
		teamRepo := do.MustInvoke[*postgres.TeamRepository](i)
		return usecase.NewCreateTeamUseCase(teamRepo), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.CreateMentorUseCase, error) {
		staffRepo := do.MustInvoke[*postgres.StaffRepository](i)
		userRepo := do.MustInvoke[*postgres.UserRepository](i)
		teamRepo := do.MustInvoke[*postgres.TeamRepository](i)
		setupTokenRepo := do.MustInvoke[*postgres.SetupTokenRepository](i)
		mentorRepo := do.MustInvoke[*postgres.MentorRepository](i)
		slackClient := do.MustInvoke[*slackinfra.Client](i)
		return usecase.NewCreateMentorUseCase(staffRepo, userRepo, teamRepo, setupTokenRepo, mentorRepo, slackClient, cfg.FrontendBaseURL), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.RegisterParticipantUseCase, error) {
		userRepo := do.MustInvoke[*postgres.UserRepository](i)
		teamRepo := do.MustInvoke[*postgres.TeamRepository](i)
		mentorRepo := do.MustInvoke[*postgres.MentorRepository](i)
		participantRepo := do.MustInvoke[*postgres.ParticipantRepository](i)
		slackClient := do.MustInvoke[*slackinfra.Client](i)
		return usecase.NewRegisterParticipantUseCase(userRepo, teamRepo, mentorRepo, participantRepo, slackClient), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.SetupPasswordUseCase, error) {
		setupTokenRepo := do.MustInvoke[*postgres.SetupTokenRepository](i)
		userRepo := do.MustInvoke[*postgres.UserRepository](i)
		hasher := do.MustInvoke[*util.PasswordHasher](i)
		database := do.MustInvoke[*sql.DB](i)
		return usecase.NewSetupPasswordUseCase(setupTokenRepo, userRepo, hasher, database), nil
	})

	// --- Handlers ---
	do.Provide(injector, func(i do.Injector) (*webhook.WebhookHandler, error) {
		uc := do.MustInvoke[*usecase.HandleProgressUseCase](i)
		return webhook.NewWebhookHandler(uc), nil
	})

	do.Provide(injector, func(i do.Injector) (*webhook.QuestionHandler, error) {
		uc := do.MustInvoke[*usecase.HandleNewQuestionUseCase](i)
		return webhook.NewQuestionHandler(uc), nil
	})

	do.Provide(injector, func(i do.Injector) (*handlers.AuthHandler, error) {
		uc := do.MustInvoke[*usecase.LoginUseCase](i)
		return handlers.NewAuthHandler(uc), nil
	})

	do.Provide(injector, func(i do.Injector) (*handlers.ProgressHandler, error) {
		uc := do.MustInvoke[*usecase.ListProgressUseCase](i)
		return handlers.NewProgressHandler(uc), nil
	})

	do.Provide(injector, func(i do.Injector) (*handlers.GitHubHandler, error) {
		svc := do.MustInvoke[*githubsvc.GitHubService](i)
		return handlers.NewGitHubHandler(svc), nil
	})

	do.Provide(injector, func(i do.Injector) (*handlers.InternalHandler, error) {
		svc := do.MustInvoke[*githubsvc.GitHubService](i)
		return handlers.NewInternalHandler(svc), nil
	})

	do.Provide(injector, func(i do.Injector) (*webhook.EventHandler, error) {
		uc := do.MustInvoke[*usecase.TriggerIssueCreationUseCase](i)
		return webhook.NewEventHandler(uc, cfg.IssueTriggerEmoji), nil
	})

	do.Provide(injector, func(i do.Injector) (*webhook.InteractionHandler, error) {
		resolveUC := do.MustInvoke[*usecase.ResolveQuestionUseCase](i)
		continueUC := do.MustInvoke[*usecase.ContinueQuestionUseCase](i)
		escalateUC := do.MustInvoke[*usecase.EscalateQuestionUseCase](i)
		return webhook.NewInteractionHandler(resolveUC, continueUC, escalateUC), nil
	})

	do.Provide(injector, func(i do.Injector) (*handlers.StaffHandler, error) {
		staffLoginUC := do.MustInvoke[*usecase.StaffLoginUseCase](i)
		createTeamUC := do.MustInvoke[*usecase.CreateTeamUseCase](i)
		teamRepo := do.MustInvoke[*postgres.TeamRepository](i)
		return handlers.NewStaffHandler(staffLoginUC, createTeamUC, teamRepo), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.ListSlackUsersUseCase, error) {
		slackClient := do.MustInvoke[*slackinfra.Client](i)
		return usecase.NewListSlackUsersUseCase(slackClient), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.ListMentorTeamsUseCase, error) {
		mentorRepo := do.MustInvoke[*postgres.MentorRepository](i)
		teamRepo := do.MustInvoke[*postgres.TeamRepository](i)
		return usecase.NewListMentorTeamsUseCase(mentorRepo, teamRepo), nil
	})

	do.Provide(injector, func(i do.Injector) (*handlers.TeamHandler, error) {
		uc := do.MustInvoke[*usecase.ListMentorTeamsUseCase](i)
		return handlers.NewTeamHandler(uc), nil
	})

	do.Provide(injector, func(i do.Injector) (*handlers.SlackHandler, error) {
		uc := do.MustInvoke[*usecase.ListSlackUsersUseCase](i)
		return handlers.NewSlackHandler(uc), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.ListTeamParticipantsUseCase, error) {
		participantRepo := do.MustInvoke[*postgres.ParticipantRepository](i)
		mentorRepo := do.MustInvoke[*postgres.MentorRepository](i)
		return usecase.NewListTeamParticipantsUseCase(participantRepo, mentorRepo), nil
	})

	do.Provide(injector, func(i do.Injector) (*handlers.ParticipantHandler, error) {
		registerUC := do.MustInvoke[*usecase.RegisterParticipantUseCase](i)
		listUC := do.MustInvoke[*usecase.ListTeamParticipantsUseCase](i)
		return handlers.NewParticipantHandler(registerUC, listUC), nil
	})

	do.Provide(injector, func(i do.Injector) (*webhook.CommandHandler, error) {
		uc := do.MustInvoke[*usecase.CreateMentorUseCase](i)
		return webhook.NewCommandHandler(uc), nil
	})

	do.Provide(injector, func(i do.Injector) (*handlers.SetupHandler, error) {
		setupPasswordUC := do.MustInvoke[*usecase.SetupPasswordUseCase](i)
		return handlers.NewSetupHandler(setupPasswordUC), nil
	})

	// --- Router ---
	do.Provide(injector, func(i do.Injector) (http.Handler, error) {
		jwtService := do.MustInvoke[*jwt.JWTService](i)
		verifier := do.MustInvoke[*pkgslack.Verifier](i)

		var corsOrigins []string
		if cfg.CORSAllowedOrigin != "" {
			for _, o := range strings.Split(cfg.CORSAllowedOrigin, ",") {
				corsOrigins = append(corsOrigins, strings.TrimSpace(o))
			}
		}

		return rest.NewRouter(rest.RouterConfig{
			AuthHandler:        do.MustInvoke[*handlers.AuthHandler](i),
			ProgressHandler:    do.MustInvoke[*handlers.ProgressHandler](i),
			GitHubHandler:      do.MustInvoke[*handlers.GitHubHandler](i),
			InternalHandler:    do.MustInvoke[*handlers.InternalHandler](i),
			StaffHandler:       do.MustInvoke[*handlers.StaffHandler](i),
			SetupHandler:       do.MustInvoke[*handlers.SetupHandler](i),
			ParticipantHandler: do.MustInvoke[*handlers.ParticipantHandler](i),
			SlackHandler:       do.MustInvoke[*handlers.SlackHandler](i),
			TeamHandler:        do.MustInvoke[*handlers.TeamHandler](i),
			CommandHandler:     do.MustInvoke[*webhook.CommandHandler](i),
			WebhookHandler:     do.MustInvoke[*webhook.WebhookHandler](i),
			QuestionHandler:    do.MustInvoke[*webhook.QuestionHandler](i),
			EventHandler:       do.MustInvoke[*webhook.EventHandler](i),
			InteractionHandler: do.MustInvoke[*webhook.InteractionHandler](i),
			JWTService:         jwtService,
			SlackVerifier:      verifier,
			InternalToken:      cfg.InternalToken,
			CORSAllowedOrigins: corsOrigins,
		}), nil
	})

	router, err := do.Invoke[http.Handler](injector)
	if err != nil {
		return nil, fmt.Errorf("failed to build router: %w", err)
	}

	return router, nil
}
