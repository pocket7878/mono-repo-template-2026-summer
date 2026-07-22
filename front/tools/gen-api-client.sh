#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FRONT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${FRONT_DIR}/.." && pwd)"

mkdir -p "${FRONT_DIR}/app/lib/backend-api/openapi"
"${FRONT_DIR}/node_modules/.bin/openapi-typescript" \
  "${REPO_ROOT}/shared/backend_api/openapi/openapi.yaml" \
  -o "${FRONT_DIR}/app/lib/backend-api/openapi/types.ts"
