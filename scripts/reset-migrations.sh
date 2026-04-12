#!/bin/bash

# Reset migrations script
# This script drops the schema_migrations table to allow a fresh start

set -e

echo "🔄 Resetting migrations..."

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Default values
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-3306}
DB_USER=${DB_USER:-root}
DB_PASSWORD=${DB_PASSWORD:-123456}
DB_NAME=${DB_NAME:-payments_ms}

echo "📊 Database configuration:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  Database: $DB_NAME"

# Check if docker is available
if command -v docker &> /dev/null; then
    # Try to find MySQL container
    CONTAINER_ID=$(docker ps --filter "expose=3306" --format "{{.ID}}" | head -n 1)
    
    if [ -n "$CONTAINER_ID" ]; then
        echo "🐳 Using Docker container: $CONTAINER_ID"
        docker exec -i $CONTAINER_ID mysql -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" <<EOF
DROP TABLE IF EXISTS schema_migrations;
SELECT 'Schema migrations table dropped successfully' as status;
EOF
        echo "✅ Migrations reset successfully!"
        exit 0
    fi
fi

echo "❌ Could not find MySQL. Please reset manually:"
echo "   DROP TABLE IF EXISTS schema_migrations;"
exit 1
