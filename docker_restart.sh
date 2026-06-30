#!/bin/sh
set -eu

# Deployment settings
CONTAINER_NAME="mikan-qb-rss"
IMAGE_NAME="ghcr.io/hanngyaku/mikan-qb-rss:latest"
HOST_PORT="18081"
DATA_DIR="/mnt/data_sda1/mikan-qb-rss"

mkdir -p "$DATA_DIR"
chown 1000:1000 "$DATA_DIR"

echo "==> Pulling latest image"
docker pull "$IMAGE_NAME"

echo "==> Replacing container"
docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true

docker run -d \
    --name "$CONTAINER_NAME" \
    --restart unless-stopped \
    -p "${HOST_PORT}:8081" \
    -e DB_PATH=/app/data/app.db \
    -e LOG_PATH=/app/data/app.log \
    -e LISTEN_ADDR=:8081 \
    -v "${DATA_DIR}:/app/data" \
    "$IMAGE_NAME"

echo "==> Running at http://<router-ip>:${HOST_PORT}"
