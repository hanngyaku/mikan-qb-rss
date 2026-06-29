#!/bin/sh
set -eu

cd "$(dirname "$0")"

mkdir -p data

echo "==> Pulling latest image"
docker compose -f docker-compose.ghcr.yml pull

echo "==> Updating and starting service"
docker compose -f docker-compose.ghcr.yml up -d --remove-orphans

echo "==> Service status"
docker compose -f docker-compose.ghcr.yml ps
