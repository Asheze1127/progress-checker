.PHONY: loadseed_db

# Seed the database with initial staff members.
# Uses DB connection settings from environment variables (defaults to local dev).
loadseed_db:
	cd backend && go run ./cmd/seed
