#!/bin/bash

# Exit on any error
set -e

# Docker container name where MySQL is running
MYSQL_CONTAINER="mysql"

# Function to run MySQL command in the container
run_mysql() {
    docker exec -i $MYSQL_CONTAINER mysql -u root -p"$MYSQL_ROOT_PASSWORD" --protocol=tcp "$@"
}

# Function to check if container is running
container_running() {
    docker ps --format '{{.Names}}' | grep -q "^$MYSQL_CONTAINER$"
}

# Check if container is running
if ! container_running; then
    echo "❌ MySQL container '$MYSQL_CONTAINER' is not running"
    echo "   Start it with: docker start $MYSQL_CONTAINER"
    exit 1
fi

# Load environment variables from .env if it exists
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

# Set default values if not set in .env
DB_NAME=${DB_NAME:-payments_ms}
MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD:-root}  # Default MySQL root password

# Check if we can connect to MySQL
echo "🔌 Testing MySQL connection..."
if ! run_mysql -e "SELECT 1" &>/dev/null; then
    echo "❌ Failed to connect to MySQL in container '$MYSQL_CONTAINER'"
    if [ -z "$MYSQL_ROOT_PASSWORD" ] || [ "$MYSQL_ROOT_PASSWORD" = "yourpassword" ]; then
        echo "   Please set the correct MYSQL_ROOT_PASSWORD in your .env file"
        echo "   Example: MYSQL_ROOT_PASSWORD=your_mysql_root_password"
    fi
    exit 1
fi

# Function to drop all tables in the database
drop_all_tables() {
    echo "🔧 Preparing to drop all tables in database: $DB_NAME"
    
    echo "🔍 Getting list of tables from database: $DB_NAME"
    
    # Get list of tables in the correct order to drop (child tables first)
    TABLES=$(run_mysql -N -e "
        SELECT TABLE_NAME 
        FROM INFORMATION_SCHEMA.TABLES 
        WHERE TABLE_SCHEMA = '$DB_NAME' 
        AND TABLE_TYPE = 'BASE TABLE'")
        
    if [ -z "$TABLES" ]; then
        echo "ℹ️  No tables found in database '$DB_NAME'"
        return 0
    fi
    
    echo "📋 Found tables: $(echo $TABLES | tr '\n' ' ')"

    # Disable foreign key checks temporarily
    run_mysql -e "SET FOREIGN_KEY_CHECKS = 0;"

    # Drop each table
    for TABLE in $TABLES; do
        echo "  🗑️  Dropping table: $TABLE"
        run_mysql -e "USE \`$DB_NAME\`; DROP TABLE IF EXISTS \`$TABLE\`;"
    done

    # Re-enable foreign key checks
    run_mysql -e "SET FOREIGN_KEY_CHECKS = 1;"

    echo "✅ All tables dropped successfully!"
}

# Main execution
main() {
    echo "🚀 Starting database cleanup for: $DB_NAME"
    
    # Confirm before proceeding
    read -p "⚠️  WARNING: This will drop ALL tables in the database. Are you sure? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        drop_all_tables
    else
        echo "❌ Operation cancelled by user"
        exit 1
    fi
}

main "$@"
