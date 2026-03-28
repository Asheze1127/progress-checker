.PHONY: dev setup_db backend frontend local-worker

# Open backend and frontend in separate VS Code terminals
dev: setup_db
	@code -n --command workbench.action.tasks.runTask "dev"

# Run DB migration and seed
setup_db:
	@$(MAKE) -C backend migrate_db loadseed_db

# Start backend server
backend:
	@cd backend && go run .

# Start frontend dev server
frontend:
	@cd web && pnpm dev

# Start local SQS worker (requires ElasticMQ running via docker-compose)
local-worker:
	@cd lambda && SQS_ENDPOINT=http://localhost:9324 pnpm run local-worker
