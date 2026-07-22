#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="${ROOT_DIR}/backend"
FRONT_DIR="${ROOT_DIR}/front"
BACKEND_LOG="${TMPDIR:-/tmp}/reference-todo-backend-e2e.log"
FRONT_LOG="${TMPDIR:-/tmp}/reference-todo-front-e2e.log"
E2E_BIN_DIR="$(mktemp -d "${TMPDIR:-/tmp}/reference-todo-e2e.XXXXXX")"

if [[ -f "${BACKEND_DIR}/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "${BACKEND_DIR}/.env"
  set +a
fi

POSTGRES_PORT="${POSTGRES_PORT:-15432}"
LOCALSTACK_PORT="${LOCALSTACK_PORT:-14566}"
BACKEND_ADDR="${BACKEND_ADDR:-127.0.0.1:8080}"
BACKEND_PORT="${BACKEND_ADDR##*:}"
FRONTEND_PORT="${FRONTEND_PORT:-5173}"
DATABASE_URL="postgres://postgres:postgres@127.0.0.1:${POSTGRES_PORT}/reference_todo_e2e?sslmode=disable"
SQS_ENDPOINT="${AWS_ENDPOINT_URL_SQS:-http://127.0.0.1:${LOCALSTACK_PORT}}"
QUEUE_URL="${OUTBOX_QUEUE_URL:-${SQS_ENDPOINT}/queue/ap-northeast-1/000000000000/reference-todo-outbox}"

cleanup() {
  [[ -n "${BACKEND_PID:-}" ]] && kill "${BACKEND_PID}" 2>/dev/null || true
  [[ -n "${FRONT_PID:-}" ]] && kill "${FRONT_PID}" 2>/dev/null || true
  [[ -n "${WORKER_PID:-}" ]] && kill "${WORKER_PID}" 2>/dev/null || true
  rm -rf "${E2E_BIN_DIR}"
}
trap cleanup EXIT

cd "${BACKEND_DIR}"
docker compose up -d postgres localstack
for _ in $(seq 1 30); do
  if docker compose exec -T postgres pg_isready -U postgres -d reference_todo >/dev/null 2>&1; then
    break
  fi
  sleep 1
done
for _ in $(seq 1 30); do
  if docker compose exec -T localstack awslocal sqs get-queue-url --queue-name reference-todo-outbox >/dev/null 2>&1; then
    break
  fi
  sleep 1
done
docker compose exec -T localstack awslocal sqs purge-queue \
  --queue-url http://localhost:4566/queue/ap-northeast-1/000000000000/reference-todo-outbox
docker compose exec -T postgres createdb -U postgres reference_todo_e2e 2>/dev/null || true
migrate -path db/migrations -database "${DATABASE_URL}" drop -f
migrate -path db/migrations -database "${DATABASE_URL}" up
go build -o "${E2E_BIN_DIR}/server" ./cmd/server
go build -o "${E2E_BIN_DIR}/worker" ./cmd/worker

DATABASE_URL="${DATABASE_URL}" \
  BACKEND_ADDR="${BACKEND_ADDR}" \
  "${E2E_BIN_DIR}/server" >"${BACKEND_LOG}" 2>&1 &
BACKEND_PID=$!

for _ in $(seq 1 30); do
  if curl -fsS "http://127.0.0.1:${BACKEND_PORT}/health" >/dev/null; then
    break
  fi
  sleep 1
done
curl -fsS "http://127.0.0.1:${BACKEND_PORT}/health" >/dev/null

DATABASE_URL="${DATABASE_URL}" \
  AWS_REGION=ap-northeast-1 \
  AWS_ACCESS_KEY_ID=test \
  AWS_SECRET_ACCESS_KEY=test \
  AWS_ENDPOINT_URL_SQS="${SQS_ENDPOINT}" \
  OUTBOX_QUEUE_URL="${QUEUE_URL}" \
  "${E2E_BIN_DIR}/worker" >"${TMPDIR:-/tmp}/reference-todo-worker-e2e.log" 2>&1 &
WORKER_PID=$!

cd "${FRONT_DIR}"
VITE_BACKEND_API_BASE_URL="http://127.0.0.1:${BACKEND_PORT}" ./node_modules/.bin/vite --host 127.0.0.1 --port "${FRONTEND_PORT}" >"${FRONT_LOG}" 2>&1 &
FRONT_PID=$!
for _ in $(seq 1 30); do
  if curl -fsS "http://127.0.0.1:${FRONTEND_PORT}/todos" >/dev/null; then
    break
  fi
  sleep 1
done
curl -fsS "http://127.0.0.1:${FRONTEND_PORT}/todos" >/dev/null
E2E_BASE_URL="http://127.0.0.1:${FRONTEND_PORT}" npm run e2e

for _ in $(seq 1 15); do
  if docker compose -f "${BACKEND_DIR}/docker-compose.yml" exec -T localstack \
    awslocal sqs receive-message \
      --queue-url http://localhost:4566/queue/ap-northeast-1/000000000000/reference-todo-outbox \
      --max-number-of-messages 10 \
      --wait-time-seconds 1 \
      --query 'Messages[0].MessageId' \
      --output text | grep -Ev '^(None|null)$'; then
    exit 0
  fi
  sleep 1
done

echo "outbox worker did not publish a message to SQS" >&2
exit 1
