.PHONY: sql-format

SQL_FORMAT_DIR ?= ../../backend/database/postgres
SQL_FORMATTER_BIN := ./node_modules/.bin/sql-formatter

sql-format:
	@if [ ! -x "$(SQL_FORMATTER_BIN)" ]; then \
		echo "SQL formatter is not installed. Run 'pnpm install' in tools/sql first."; \
		exit 1; \
	elif [ ! -d "$(SQL_FORMAT_DIR)" ]; then \
		echo "Directory $(SQL_FORMAT_DIR) does not exist. Skipping format."; \
	elif ! find "$(SQL_FORMAT_DIR)" -name "*.sql" -print -quit | grep -q .; then \
		echo "No SQL files found in $(SQL_FORMAT_DIR)/. Skipping format."; \
	else \
		find "$(SQL_FORMAT_DIR)" -name "*.sql" -print0 | xargs -0 -n 1 "$(SQL_FORMATTER_BIN)" --fix --language postgresql; \
	fi
