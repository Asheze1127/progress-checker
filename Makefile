.PHONY: dev setup_db backend frontend

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
