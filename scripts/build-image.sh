#!/bin/bash

# Script para construir la imagen Docker con versionado
# Uso: ./scripts/build-image.sh [version]
# Ejemplo: ./scripts/build-image.sh 1.0.1

set -e

# Cargar variables del archivo .env.docker
if [ -f .env.docker ]; then
    export $(cat .env.docker | grep -v '^#' | xargs)
fi

# Si se proporciona una versión como argumento, usarla
if [ -n "$1" ]; then
    IMAGE_VERSION=$1
fi

# Valores por defecto
IMAGE_NAME=${IMAGE_NAME:-payments-ms}
IMAGE_VERSION=${IMAGE_VERSION:-latest}

echo "=========================================="
echo "Building Docker Image"
echo "=========================================="
echo "Image Name: ${IMAGE_NAME}"
echo "Version: ${IMAGE_VERSION}"
echo "=========================================="

# Construir la imagen con la versión especificada
docker build -t ${IMAGE_NAME}:${IMAGE_VERSION} .

# También etiquetar como latest
docker tag ${IMAGE_NAME}:${IMAGE_VERSION} ${IMAGE_NAME}:latest

echo "=========================================="
echo "✅ Image built successfully!"
echo "=========================================="
echo "Images created:"
echo "  - ${IMAGE_NAME}:${IMAGE_VERSION}"
echo "  - ${IMAGE_NAME}:latest"
echo "=========================================="
