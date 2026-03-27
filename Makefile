.PHONY: dev setup_db

# Start everything: DB setup + backend + frontend
dev: setup_db
	@echo "Starting backend and frontend..."
	@cd backend && go run . & \
	cd web && pnpm dev & \
	wait

# Run DB migration and seed
setup_db:
	@$(MAKE) -C backend migrate_db loadseed_db
