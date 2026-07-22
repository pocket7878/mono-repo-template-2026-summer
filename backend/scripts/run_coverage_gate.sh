#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
THRESHOLD="${COVERAGE_THRESHOLD:-85.0}"
COVERPROFILE="$(mktemp "${TMPDIR:-/tmp}/reference-todo-coverage.XXXXXX")"
trap 'rm -f "${COVERPROFILE}"' EXIT

cd "${BACKEND_DIR}"
if [[ -z "${POSTGRES_TEST_DSN:-}" ]]; then
  docker compose up -d postgres
  for _ in $(seq 1 30); do
    if docker compose exec -T postgres pg_isready -U postgres -d reference_todo >/dev/null 2>&1; then
      break
    fi
    sleep 1
  done
  docker compose exec -T postgres pg_isready -U postgres -d reference_todo >/dev/null
  export POSTGRES_TEST_DSN="postgres://postgres:postgres@127.0.0.1:${POSTGRES_PORT:-15432}/reference_todo?sslmode=disable"
fi

POSTGRES_TEST_DSN="${POSTGRES_TEST_DSN}" \
  go test -p 1 -coverprofile="${COVERPROFILE}" ./...
go run ./cmd/coverage-gate -coverprofile "${COVERPROFILE}" -threshold "${THRESHOLD}"
