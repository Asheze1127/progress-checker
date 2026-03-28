.PHONY: migrate_db loadseed_db

# DB connection settings (override with environment variables)
DATABASE_HOST ?= localhost
DATABASE_PORT ?= 5432
DATABASE_NAME ?= progress_checker
DATABASE_USER ?= postgres
DATABASE_PASS ?= postgres

# Apply schema to the database using psqldef.
# Installs psqldef automatically if not found.
GOBIN := $(shell go env GOPATH)/bin

migrate_db:
	@command -v psqldef >/dev/null 2>&1 || { \
		echo "Installing psqldef..."; \
		go install github.com/sqldef/sqldef/cmd/psqldef@latest; \
	}
	PGPASSWORD=$(DATABASE_PASS) PATH="$(GOBIN):$(PATH)" psqldef \
		-h $(DATABASE_HOST) \
		-p $(DATABASE_PORT) \
		-U $(DATABASE_USER) \
		$(DATABASE_NAME) < database/postgres/schema.sql

# Seed the database with initial staff members.
# Run migrate_db first to ensure the schema is up to date.
loadseed_db: migrate_db
	go run ./cmd/seed
