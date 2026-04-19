#!/usr/bin/env bash
set -euo pipefail

IMAGE="${1:-${IMAGE:-}}"
REPO_DIR="${2:-${REPO_DIR:-/opt/new-api-deploy}}"
ENV_FILE="$REPO_DIR/deploy/.env.prod"
COMPOSE_FILE="$REPO_DIR/deploy/compose.prod.yml"
STATUS_URL="${STATUS_URL:-http://127.0.0.1:3000/api/status}"

compose() {
  if docker compose version >/dev/null 2>&1; then
    docker compose "$@"
  else
    docker-compose "$@"
  fi
}

if [[ -z "$IMAGE" ]]; then
  echo "usage: remote-deploy-prod.sh <image>"
  exit 1
fi

if [[ ! -d "$REPO_DIR" ]]; then
  echo "repo dir not found: $REPO_DIR"
  exit 1
fi

if [[ ! -f "$ENV_FILE" ]]; then
  echo "env file not found: $ENV_FILE"
  exit 1
fi

if [[ ! -f "$COMPOSE_FILE" ]]; then
  echo "compose file not found: $COMPOSE_FILE"
  exit 1
fi

cd "$REPO_DIR"

if grep -q '^NEW_API_IMAGE=' "$ENV_FILE"; then
  sed -i "s#^NEW_API_IMAGE=.*#NEW_API_IMAGE=$IMAGE#" "$ENV_FILE"
else
  printf '\nNEW_API_IMAGE=%s\n' "$IMAGE" >> "$ENV_FILE"
fi

compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" pull new-api
compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d new-api

for _ in $(seq 1 30); do
  if curl -fsS --max-time 5 "$STATUS_URL" | grep -q '"success":true'; then
    echo "deployed image: $IMAGE"
    exit 0
  fi
  sleep 2
done

echo "health check failed: $STATUS_URL" >&2
compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps >&2

exit 1
