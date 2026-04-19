#!/usr/bin/env bash
set -euo pipefail

IMAGE="${1:-${IMAGE:-}}"
REPO_DIR="${2:-${REPO_DIR:-/root/new-api-deploy}}"
ENV_FILE="$REPO_DIR/deploy/.env.prod"
COMPOSE_FILE="$REPO_DIR/deploy/compose.prod.yml"

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

git fetch origin main
git checkout main
git pull --ff-only origin main

sed -i "s#^NEW_API_IMAGE=.*#NEW_API_IMAGE=$IMAGE#" "$ENV_FILE"

docker-compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" pull new-api
docker-compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d new-api

curl -fsS http://127.0.0.1:3000/api/status | grep -q '"success":true'

echo "deployed image: $IMAGE"
