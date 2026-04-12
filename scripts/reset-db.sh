#!/bin/bash

# Database reset script
# This script drops and recreates the database without any tables using docker exec

set -e

echo "🔄 Resetting database..."

# Default values
DB_NAME=${DB_NAME:-payments_ms}
MYSQL_CONTAINER=mysql
MYSQL_USER=root
MYSQL_PASSWORD=123456

echo "📊 Database configuration:"
echo "  Container: $MYSQL_CONTAINER"
echo "  User: $MYSQL_USER"
echo "  Database: $DB_NAME"

# Function to run MySQL commands inside the container
run_mysql() {
    echo "Running: $*"
    docker exec -i $MYSQL_CONTAINER mysql -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "$*"
}

# Check if MySQL container is running
if ! docker ps | grep -q $MYSQL_CONTAINER; then
    echo "❌ MySQL container '$MYSQL_CONTAINER' is not running"
    echo "   Start it with: docker start $MYSQL_CONTAINER"
    exit 1
fi

# Check if we can connect to MySQL
if ! docker exec $MYSQL_CONTAINER mysql -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "SELECT 1" &>/dev/null; then
    echo "❌ Failed to connect to MySQL. Please check your credentials."
    exit 1
fi

# Drop and recreate the database
echo "🗑️  Dropping database '$DB_NAME' if it exists..."
run_mysql "DROP DATABASE IF EXISTS \`$DB_NAME\`;"

echo "🆕 Creating new database '$DB_NAME'..."
run_mysql "CREATE DATABASE \`$DB_NAME\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

echo "✅ Database '$DB_NAME' has been reset successfully!"
echo "   You can now run migrations with: ./scripts/init-db.sh"
