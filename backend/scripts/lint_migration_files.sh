#!/bin/sh

set -eu

migration_dir="${1:-db/migrations}"

if [ ! -d "${migration_dir}" ]; then
  echo "migration directory not found: ${migration_dir}" >&2
  exit 1
fi

status=0

for migration_file in "${migration_dir}"/*.sql; do
  [ -e "${migration_file}" ] || continue
  filename="${migration_file##*/}"
  if ! printf '%s\n' "${filename}" | grep -Eq '^[0-9]{14}_[a-z0-9_]+\.(up|down)\.sql$'; then
    echo "invalid migration filename: ${migration_file}" >&2
    status=1
  fi
done

for up_file in "${migration_dir}"/*.up.sql; do
  [ -e "${up_file}" ] || continue
  down_file="${up_file%.up.sql}.down.sql"
  if [ ! -f "${down_file}" ]; then
    echo "missing down migration for ${up_file}" >&2
    status=1
  fi
done

for down_file in "${migration_dir}"/*.down.sql; do
  [ -e "${down_file}" ] || continue
  up_file="${down_file%.down.sql}.up.sql"
  if [ ! -f "${up_file}" ]; then
    echo "missing up migration for ${down_file}" >&2
    status=1
  fi
done

if grep -Eiq '(^|[^[:alnum:]_])(create|drop)[[:space:]]+extension([^[:alnum:]_]|$)' "${migration_dir}"/*.sql; then
  echo "application migrations must not create or drop PostgreSQL extensions" >&2
  status=1
fi

if grep -Eiq '(^|[^[:alnum:]_])(grant|revoke)([^[:alnum:]_]|$)|alter[[:space:]]+[^;]+[[:space:]]+owner[[:space:]]+to' "${migration_dir}"/*.sql; then
  echo "application migrations must not change database privileges or object ownership" >&2
  status=1
fi

exit "${status}"
