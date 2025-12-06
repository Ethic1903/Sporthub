#!/usr/bin/env bash
set -euo pipefail

DATABASE_URL="${DATABASE_URL:-postgres://postgres:postgres@localhost:5432/sporthub?sslmode=disable}"
PSQL_BIN="${PSQL:-psql}"
SERVICES=(identity facility booking)

if ! command -v "$PSQL_BIN" >/dev/null 2>&1; then
  echo "psql command not found (looked for '$PSQL_BIN')" >&2
  exit 1
fi

echo "==> applying migrations to $DATABASE_URL"
for svc in "${SERVICES[@]}"; do
  dir="services/$svc/migrations"
  if [[ ! -d "$dir" ]]; then
    continue
  fi
  shopt -s nullglob
  files=("$dir"/*.sql)
  shopt -u nullglob
  if (( ${#files[@]} == 0 )); then
    continue
  fi
  echo "  -> $svc"
  for file in "${files[@]}"; do
    echo "     applying $(basename "$file")"
    "$PSQL_BIN" "$DATABASE_URL" -f "$file"
  done
done

echo "==> migrations complete"
