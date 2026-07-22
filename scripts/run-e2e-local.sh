#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="${ROOT_DIR}/backend"
FRONT_DIR="${ROOT_DIR}/front"
BACKEND_LOG="${TMPDIR:-/tmp}/reference-todo-backend-e2e.log"
FRONT_LOG="${TMPDIR:-/tmp}/reference-todo-front-e2e.log"

cleanup() {
  [[ -n "${BACKEND_PID:-}" ]] && kill "${BACKEND_PID}" 2>/dev/null || true
  [[ -n "${FRONT_PID:-}" ]] && kill "${FRONT_PID}" 2>/dev/null || true
}
trap cleanup EXIT

cd "${BACKEND_DIR}"
docker compose up -d postgres
for _ in $(seq 1 30); do
  if docker compose exec -T postgres pg_isready -U postgres -d reference_todo >/dev/null 2>&1; then
    break
  fi
  sleep 1
done
docker compose exec -T postgres psql -U postgres -d reference_todo -c "DROP TABLE IF EXISTS job_dead_letters, job_outbox_events, todos CASCADE;"
docker compose exec -T postgres psql -U postgres -d reference_todo < db/migrations/20260722000000_create_todos.up.sql
docker compose exec -T postgres psql -U postgres -d reference_todo < db/migrations/20260722010000_create_job_outbox_events.up.sql

DATABASE_URL="postgres://postgres:postgres@127.0.0.1:15432/reference_todo?sslmode=disable" \
  BACKEND_ADDR=127.0.0.1:8080 \
  go run ./cmd/server >"${BACKEND_LOG}" 2>&1 &
BACKEND_PID=$!

for _ in $(seq 1 30); do
  if curl -fsS http://127.0.0.1:8080/health >/dev/null; then
    break
  fi
  sleep 1
done
curl -fsS http://127.0.0.1:8080/health >/dev/null

cd "${FRONT_DIR}"
VITE_BACKEND_API_BASE_URL=http://127.0.0.1:8080 npm run dev -- --host 127.0.0.1 --port 5173 >"${FRONT_LOG}" 2>&1 &
FRONT_PID=$!
for _ in $(seq 1 30); do
  if curl -fsS http://127.0.0.1:5173/todos >/dev/null; then
    break
  fi
  sleep 1
done
curl -fsS http://127.0.0.1:5173/todos >/dev/null
npm run e2e
