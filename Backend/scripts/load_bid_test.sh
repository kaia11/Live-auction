#!/bin/sh

set -eu

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
MODE="${1:-single}"
CONCURRENCY="${CONCURRENCY:-5}"
REQUESTS_PER_WORKER="${REQUESTS_PER_WORKER:-20}"
ROOM_IDS="${ROOM_IDS:-room-001}"
SESSION_ID="${SESSION_ID:-session-001}"
ITEM_ID="${ITEM_ID:-item-001}"
START_PRICE="${START_PRICE:-140}"

TOKEN="$(curl -sS -X POST "$BASE_URL/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"phone":"viewer","password":"demo"}' | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')"

if [ -z "$TOKEN" ]; then
  echo "failed to get token"
  exit 1
fi

run_worker() {
  worker_id="$1"
  request_count="$2"
  room_id="$3"

  i=0
  while [ "$i" -lt "$request_count" ]; do
    price=$((START_PRICE + worker_id * 100 + i))
    request_id="load-${MODE}-${worker_id}-${i}"
    curl -sS -X POST "$BASE_URL/bids" \
      -H "Authorization: Bearer $TOKEN" \
      -H 'Content-Type: application/json' \
      -d "{\"roomId\":\"$room_id\",\"sessionId\":\"$SESSION_ID\",\"itemId\":\"$ITEM_ID\",\"bidPrice\":$price,\"requestId\":\"$request_id\"}" >/dev/null || true
    i=$((i + 1))
  done
}

worker=0
set -- $(printf "%s" "$ROOM_IDS" | tr ',' ' ')
ROOM_COUNT=$#

while [ "$worker" -lt "$CONCURRENCY" ]; do
  if [ "$MODE" = "multi" ] && [ "$ROOM_COUNT" -gt 0 ]; then
    index=$((worker % ROOM_COUNT + 1))
    eval room_id=\${$index}
  else
    room_id="${1:-room-001}"
  fi
  run_worker "$worker" "$REQUESTS_PER_WORKER" "$room_id" &
  worker=$((worker + 1))
done

wait
echo "load test completed mode=$MODE concurrency=$CONCURRENCY requests_per_worker=$REQUESTS_PER_WORKER"
