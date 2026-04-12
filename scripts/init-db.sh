#!/bin/bash

# Database initialization script
# This script creates the database and runs migrations

set -e

echo "🚀 Initializing database..."

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
echo "  User: $DB_USER"
echo "  Database: $DB_NAME"

# Wait for MySQL to be ready
echo "⏳ Waiting for MySQL to be ready..."
max_attempts=30
attempt=0

while ! mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -e "SELECT 1" &> /dev/null; do
    attempt=$((attempt + 1))
    if [ $attempt -ge $max_attempts ]; then
        echo "❌ MySQL is not available after $max_attempts attempts"
        exit 1
    fi
    echo "  Attempt $attempt/$max_attempts..."
    sleep 2
done

echo "✅ MySQL is ready!"

# Run migrations
echo "📝 Running migrations..."
mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" < migrations/001_initial_schema.sql

echo "✅ Database initialized successfully!"
echo ""
echo "🎉 You can now start the application with:"
echo "   go run cmd/api/main.go"
echo "   or"
echo "   make run"
