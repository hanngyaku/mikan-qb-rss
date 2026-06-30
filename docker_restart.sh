#!/bin/sh

# Deployment settings
CONTAINER_NAME="mikan-qb-rss"
IMAGE_NAME="ghcr.io/hanngyaku/mikan-qb-rss:latest"
HOST_PORT="18082"
DATA_DIR="/mnt/data_sda1/mikan-qb-rss"

mkdir -p "$DATA_DIR"
chown 1000:1000 "$DATA_DIR"

echo "==> Pulling latest image"
docker pull "$IMAGE_NAME" || exit 1

echo "==> Replacing container"
docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true

PORT_OWNER="$(docker ps --filter "publish=${HOST_PORT}" --format '{{.Names}}' | head -n 1)"
if [ -n "$PORT_OWNER" ]; then
    echo "ERROR: port ${HOST_PORT} is already used by container: ${PORT_OWNER}"
    echo "Change HOST_PORT or remove the old container."
    exit 1
fi

docker run -d \
    --name "$CONTAINER_NAME" \
    --restart unless-stopped \
    -p "${HOST_PORT}:8081" \
    -e DB_PATH=/app/data/app.db \
    -e LOG_PATH=/app/data/app.log \
    -e LISTEN_ADDR=:8081 \
    -v "${DATA_DIR}:/app/data" \
    "$IMAGE_NAME" || exit 1

echo "==> Running at http://<router-ip>:${HOST_PORT}"
