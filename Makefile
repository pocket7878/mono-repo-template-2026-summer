REPO_ROOT := $(CURDIR)
API_DIR := $(REPO_ROOT)/shared/backend_api
BACKEND_DIR := $(REPO_ROOT)/backend
FRONT_DIR := $(REPO_ROOT)/front

.PHONY: setup gen check-gen db-up db-migrate backend-test frontend-test e2e check

setup:
	cd $(API_DIR) && npm ci
	cd $(FRONT_DIR) && npm ci
	cd $(BACKEND_DIR) && go mod download

gen:
	cd $(API_DIR) && npx tsp compile .
	cd $(BACKEND_DIR) && go generate ./internal/handler/api
	cd $(FRONT_DIR) && bash ./tools/gen-api-client.sh

check-gen:
	python3 ./scripts/check_generated_file_sync.py

db-up:
	cd $(BACKEND_DIR) && docker compose up -d postgres

db-migrate: db-up
	cd $(BACKEND_DIR) && docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U postgres -d reference_todo < db/migrations/20260722000000_create_todos.up.sql
	cd $(BACKEND_DIR) && docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U postgres -d reference_todo < db/migrations/20260722010000_create_job_outbox_events.up.sql

backend-test:
	cd $(BACKEND_DIR) && go run ./cmd/backend-lint ./... && ./scripts/run_coverage_gate.sh

frontend-test:
	cd $(FRONT_DIR) && npm run test

e2e:
	./scripts/run-e2e-local.sh

check: check-gen backend-test frontend-test
