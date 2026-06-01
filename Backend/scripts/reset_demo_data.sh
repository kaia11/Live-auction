#!/bin/sh

set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)"

cd "$ROOT_DIR"

if command -v docker >/dev/null 2>&1; then
  DOCKER_BIN="docker"
elif command -v docker.exe >/dev/null 2>&1; then
  DOCKER_BIN="docker.exe"
else
  echo "docker or docker.exe is required to reset demo data" >&2
  exit 1
fi

compose() {
  "$DOCKER_BIN" compose "$@"
}

compose up -d mysql redis
compose stop backend || true
compose exec -T mysql mysql -uroot -ppassword auction_live < Backend/mysql/schema.sql
compose exec -T mysql mysql -uroot -ppassword auction_live < Backend/mysql/seed.sql
compose exec -T redis redis-cli FLUSHALL
compose up -d backend

echo "Demo data reset completed."
