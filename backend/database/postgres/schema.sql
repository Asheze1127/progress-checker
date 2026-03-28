-- PostgreSQL schema for progress-checker
-- Compatible with psqldef (declarative schema management)

-- =============================================================================
-- ENUM types
-- =============================================================================
CREATE TYPE user_role AS ENUM ('participant', 'mentor');
CREATE TYPE slack_channel_purpose AS ENUM ('progress', 'question', 'notice');
CREATE TYPE progress_phase AS ENUM ('idea', 'design', 'coding', 'testing', 'demo');
CREATE TYPE question_status AS ENUM ('open', 'in_progress', 'awaiting_user', 'assigned_mentor', 'resolved');

-- =============================================================================
-- Trigger function: auto-update updated_at on row modification
-- =============================================================================
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- Users: participants and mentors authenticate via Slack
-- =============================================================================
CREATE TABLE users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    slack_user_id VARCHAR     NOT NULL UNIQUE,
    name          VARCHAR     NOT NULL,
    email         VARCHAR     NOT NULL,
    role          user_role   NOT NULL,
    password_hash VARCHAR     NOT NULL DEFAULT '!',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_users_email ON users (email);

CREATE TRIGGER set_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- =============================================================================
-- Teams: groups of participants and mentors
-- =============================================================================
CREATE TABLE teams (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR     NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER set_teams_updated_at
    BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- =============================================================================
-- Slack channels: linked to a team, identified by Slack's own channel ID
-- =============================================================================
CREATE TABLE slack_channels (
    id         VARCHAR     PRIMARY KEY,  -- Slack channel ID (e.g. C01ABCDEF)
    team_id    UUID        NOT NULL REFERENCES teams (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_slack_channels_team_id ON slack_channels (team_id);

-- =============================================================================
-- Channel purpose assignments: each channel can serve one or more purposes
-- =============================================================================
CREATE TABLE channel_purpose_assignments (
    slack_channel_id VARCHAR NOT NULL REFERENCES slack_channels (id) ON DELETE CASCADE,
    purpose          slack_channel_purpose NOT NULL,
    created_at       TIMESTAMPTZ          NOT NULL DEFAULT now(),
    UNIQUE (slack_channel_id, purpose)
);

CREATE INDEX idx_channel_purpose_assignments_slack_channel_id ON channel_purpose_assignments (slack_channel_id);

-- =============================================================================
-- Participants: a user with role='participant' assigned to exactly one team
-- =============================================================================
CREATE TABLE participants (
    user_id    UUID        PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    team_id    UUID        NOT NULL REFERENCES teams (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_participants_team_id ON participants (team_id);

-- =============================================================================
-- Mentors: a user with role='mentor'
-- =============================================================================
CREATE TABLE mentors (
    user_id    UUID        PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- =============================================================================
-- Mentor-team assignments: many-to-many between mentors and teams
-- =============================================================================
CREATE TABLE mentor_team_assignments (
    mentor_user_id UUID        NOT NULL REFERENCES mentors (user_id) ON DELETE CASCADE,
    team_id        UUID        NOT NULL REFERENCES teams (id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (mentor_user_id, team_id)
);

CREATE INDEX idx_mentor_team_assignments_team_id ON mentor_team_assignments (team_id);

-- =============================================================================
-- Staff: administrative users (Slack link is optional)
-- =============================================================================
CREATE TABLE staff (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    slack_user_id VARCHAR     UNIQUE,
    name          VARCHAR     NOT NULL,
    email         VARCHAR     NOT NULL UNIQUE,
    password_hash VARCHAR     NOT NULL DEFAULT '!',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER set_staff_updated_at
    BEFORE UPDATE ON staff
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- =============================================================================
-- Setup tokens: one-time password setup links for new users
-- =============================================================================
CREATE TABLE setup_tokens (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash VARCHAR     NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_setup_tokens_user_id ON setup_tokens (user_id);

-- =============================================================================
-- Progress logs: one log per participant submission
-- =============================================================================
CREATE TABLE progress_logs (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    participant_id UUID        NOT NULL REFERENCES participants (user_id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_progress_logs_participant_id ON progress_logs (participant_id);

-- =============================================================================
-- Progress bodies: individual phase entries within a progress log
-- =============================================================================
CREATE TABLE progress_bodies (
    id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    progress_log_id UUID          NOT NULL REFERENCES progress_logs (id) ON DELETE CASCADE,
    phase           progress_phase NOT NULL,
    sos             BOOLEAN       NOT NULL DEFAULT FALSE,
    comment         TEXT,
    submitted_at    TIMESTAMPTZ   NOT NULL
);

CREATE UNIQUE INDEX idx_progress_bodies_log_phase ON progress_bodies (progress_log_id, phase);
CREATE INDEX idx_progress_bodies_progress_log_id ON progress_bodies (progress_log_id);

-- =============================================================================
-- Questions: participant questions routed to mentors via Slack threads
-- =============================================================================
CREATE TABLE questions (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    participant_id   UUID        NOT NULL REFERENCES participants (user_id) ON DELETE CASCADE,
    title            VARCHAR     NOT NULL,
    slack_channel_id VARCHAR     NOT NULL REFERENCES slack_channels (id) ON DELETE RESTRICT,
    status           question_status NOT NULL,
    slack_thread_ts  VARCHAR     NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_questions_participant_id ON questions (participant_id);
CREATE INDEX idx_questions_slack_channel_id ON questions (slack_channel_id);
CREATE INDEX idx_questions_status ON questions (status);
CREATE INDEX idx_questions_channel_thread ON questions (slack_channel_id, slack_thread_ts);

CREATE TRIGGER set_questions_updated_at
    BEFORE UPDATE ON questions
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- =============================================================================
-- Question-mentor assignments: many-to-many between questions and mentors
-- =============================================================================
CREATE TABLE question_mentor_assignments (
    question_id    UUID        NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
    mentor_user_id UUID        NOT NULL REFERENCES mentors (user_id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (question_id, mentor_user_id)
);

CREATE INDEX idx_question_mentor_assignments_mentor_user_id ON question_mentor_assignments (mentor_user_id);

-- =============================================================================
-- Idempotency keys: prevent duplicate processing of Slack events / API calls
-- =============================================================================
CREATE TABLE idempotency_keys (
    key        VARCHAR     PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_idempotency_keys_expires_at ON idempotency_keys (expires_at);

-- =============================================================================
-- Team GitHub repositories: store repo URLs and encrypted PATs (Issue #69)
-- =============================================================================
CREATE TABLE team_github_repositories (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id         UUID        NOT NULL REFERENCES teams (id) ON DELETE CASCADE,
    github_repo_url VARCHAR     NOT NULL,
    encrypted_pat   VARCHAR     NOT NULL,
    UNIQUE (team_id, github_repo_url),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_team_github_repositories_team_id ON team_github_repositories (team_id);

CREATE TRIGGER set_team_github_repositories_updated_at
    BEFORE UPDATE ON team_github_repositories
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
