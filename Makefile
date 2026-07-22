REPO_ROOT := $(CURDIR)
API_DIR := $(REPO_ROOT)/shared/backend_api
BACKEND_DIR := $(REPO_ROOT)/backend
FRONT_DIR := $(REPO_ROOT)/front

.PHONY: setup setup-local-worktree-env gen check-gen check-sqlc migration-lint db-up db-migrate backend-test frontend-test e2e check

setup:
	cd $(API_DIR) && npm ci
	cd $(FRONT_DIR) && npm ci
	cd $(BACKEND_DIR) && go mod download

setup-local-worktree-env:
	./scripts/setup-local-worktree-env.sh

gen:
	cd $(API_DIR) && npx tsp compile .
	cd $(BACKEND_DIR) && go generate ./internal/handler/api
	cd $(BACKEND_DIR) && sqlc generate
	cd $(FRONT_DIR) && bash ./tools/gen-api-client.sh

check-gen:
	python3 ./scripts/check_generated_file_sync.py
	$(MAKE) check-sqlc

check-sqlc:
	cd $(BACKEND_DIR) && sqlc vet && sqlc generate
	git diff --exit-code -- backend/internal/repository/internal/db
	test -z "$$(git ls-files --others --exclude-standard -- backend/internal/repository/internal/db)"

migration-lint:
	cd $(BACKEND_DIR) && ./scripts/lint_migration_files.sh ./db/migrations

db-up:
	cd $(BACKEND_DIR) && docker compose up -d postgres localstack

db-migrate: db-up
	cd $(BACKEND_DIR) && until docker compose exec -T postgres pg_isready -U postgres -d reference_todo >/dev/null 2>&1; do sleep 1; done
	cd $(BACKEND_DIR) && set -a; [ ! -f .env ] || . ./.env; set +a; migrate -path db/migrations -database "$${DATABASE_URL:-postgres://postgres:postgres@127.0.0.1:15432/reference_todo?sslmode=disable}" up

backend-test:
	cd $(BACKEND_DIR) && go run ./cmd/backend-lint ./... && ./scripts/run_coverage_gate.sh

frontend-test:
	cd $(FRONT_DIR) && npm run test

e2e:
	./scripts/run-e2e-local.sh

check: check-gen backend-test frontend-test
