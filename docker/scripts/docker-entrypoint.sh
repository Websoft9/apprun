#!/bin/sh
set -e

# =============================================================================
# AppRun Docker Entrypoint
# Handles database migration before starting the application
# =============================================================================

# Configuration
DB_WAIT_TIMEOUT=${DB_WAIT_TIMEOUT:-30}
MIGRATIONS_DIR=${MIGRATIONS_DIR:-/app/migrations}

# Build database URL from environment variables
build_db_url() {
    local host=${DATABASE_HOST:-localhost}
    local port=${DATABASE_PORT:-5432}
    local user=${DATABASE_USER:-apprun}
    local pass=${DATABASE_PASSWORD:-}
    local name=${DATABASE_DB_NAME:-apprun}
    echo "postgres://${user}:${pass}@${host}:${port}/${name}?sslmode=disable"
}

# Wait for database to be ready
wait_for_db() {
    local url="$1"
    local timeout=$DB_WAIT_TIMEOUT
    local count=0
    
    echo "⏳ Waiting for database (timeout: ${timeout}s)..."
    
    # Extract host and port from URL for nc check
    local host=$(echo "$url" | sed -E 's|.*@([^:]+):([0-9]+)/.*|\1|')
    local port=$(echo "$url" | sed -E 's|.*@([^:]+):([0-9]+)/.*|\2|')
    
    while ! nc -z "$host" "$port" 2>/dev/null; do
        count=$((count + 1))
        if [ $count -ge $timeout ]; then
            echo "❌ Database connection timeout after ${timeout}s"
            exit 1
        fi
        echo "   Waiting for $host:$port... ($count/${timeout})"
        sleep 1
    done
    
    echo "✅ Database is reachable"
}

# Run database migrations
run_migrations() {
    local url="$1"
    
    echo "🔄 Running database migrations..."
    
    if [ ! -d "$MIGRATIONS_DIR" ]; then
        echo "⚠️  Migrations directory not found: $MIGRATIONS_DIR"
        echo "   Skipping migrations"
        return 0
    fi
    
    # Check if there are any migration files
    if [ -z "$(ls -A $MIGRATIONS_DIR/*.sql 2>/dev/null)" ]; then
        echo "ℹ️  No migration files found, skipping"
        return 0
    fi
    
    # Run Atlas migrations
    if /usr/local/bin/atlas migrate apply \
        --dir "file://$MIGRATIONS_DIR" \
        --url "$url" \
        --allow-dirty; then
        echo "✅ Migrations applied successfully"
    else
        echo "❌ Migration failed!"
        exit 1
    fi
}

# Main entrypoint logic
main() {
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "🐳 AppRun Container Starting..."
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Build database URL
    DB_URL=$(build_db_url)
    
    # Check if auto-migration is enabled (consistent with DATABASE_AUTO_MIGRATE in Go config)
    if [ "${DATABASE_AUTO_MIGRATE:-false}" = "true" ]; then
        echo "📌 DATABASE_AUTO_MIGRATE=true"
        wait_for_db "$DB_URL"
        run_migrations "$DB_URL"
    else
        echo "📌 DATABASE_AUTO_MIGRATE=false (migrations skipped)"
        echo "   Set DATABASE_AUTO_MIGRATE=true to enable startup migrations"
    fi
    
    echo ""
    echo "🚀 Starting AppRun server..."
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Execute the main application
    exec "$@"
}

# Run main with all arguments
main "$@"
