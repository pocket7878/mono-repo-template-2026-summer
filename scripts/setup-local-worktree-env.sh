#!/bin/sh

set -eu

repository_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)"
environment_file="${repository_root}/backend/.env"

if [ -f "${environment_file}" ] && [ "${1:-}" != "--force" ]; then
  echo "${environment_file} already exists; use --force to replace it" >&2
  exit 1
fi

checksum="$(printf '%s' "${repository_root}" | cksum | awk '{print $1}')"
offset=$((checksum % 1000))
postgres_port=$((20000 + offset))
localstack_port=$((22000 + offset))
backend_port=$((24000 + offset))
frontend_port=$((26000 + offset))
project_suffix="$(printf '%s' "${checksum}" | cut -c1-8)"

{
  printf 'COMPOSE_PROJECT_NAME=reference-todo-%s\n' "${project_suffix}"
  printf 'POSTGRES_PORT=%s\n' "${postgres_port}"
  printf 'LOCALSTACK_PORT=%s\n' "${localstack_port}"
  printf 'DATABASE_URL=postgres://postgres:postgres@127.0.0.1:%s/reference_todo?sslmode=disable\n' "${postgres_port}"
  printf 'BACKEND_ADDR=127.0.0.1:%s\n' "${backend_port}"
  printf 'FRONTEND_PORT=%s\n' "${frontend_port}"
	printf 'FRONTEND_ORIGIN=http://127.0.0.1:%s\n' "${frontend_port}"
  printf 'VITE_BACKEND_API_BASE_URL=http://127.0.0.1:%s\n' "${backend_port}"
  printf 'AWS_REGION=ap-northeast-1\n'
  printf 'AWS_ACCESS_KEY_ID=test\n'
  printf 'AWS_SECRET_ACCESS_KEY=test\n'
  printf 'AWS_ENDPOINT_URL_SQS=http://127.0.0.1:%s\n' "${localstack_port}"
  printf 'OUTBOX_QUEUE_URL=http://127.0.0.1:%s/queue/ap-northeast-1/000000000000/reference-todo-outbox\n' "${localstack_port}"
} >"${environment_file}"

echo "created ${environment_file}"
echo "compose project: reference-todo-${project_suffix}"
echo "postgres=${postgres_port} localstack=${localstack_port} backend=${backend_port} frontend=${frontend_port}"
