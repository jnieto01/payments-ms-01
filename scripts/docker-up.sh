#!/bin/bash

# Script para levantar el contenedor con docker-compose
# Uso: ./scripts/docker-up.sh [version]
# Ejemplo: ./scripts/docker-up.sh 1.0.1

set -e

# Si se proporciona una versión como argumento, exportarla
if [ -n "$1" ]; then
    export IMAGE_VERSION=$1
fi

# Cargar variables del archivo .env.docker si no se proporcionó versión
if [ -z "$IMAGE_VERSION" ] && [ -f .env.docker ]; then
    export $(cat .env.docker | grep -v '^#' | xargs)
fi

echo "=========================================="
echo "Starting Docker Compose"
echo "=========================================="
echo "Image: ${IMAGE_NAME:-user-ms}:${IMAGE_VERSION:-latest}"
echo "=========================================="

# Levantar con docker-compose
docker-compose --env-file .env.docker up --build -d

echo "=========================================="
echo "✅ Container started successfully!"
echo "=========================================="
echo "View logs: docker-compose logs -f"
echo "Stop: docker-compose down"
echo "=========================================="
