#!/bin/bash
# Test script for verifying reset → sync workflow
# Usage: ./scripts/test-reset-sync-workflow.sh

set -e

echo "🧪 Testing reset → sync workflow"
echo "================================"
echo ""

# Build apprun if needed
if [ ! -f "./core/bin/apprun" ]; then
    echo "📦 Building apprun..."
    cd core && go build -o bin/apprun main.go && cd ..
fi

echo "Step 1: Check current status"
echo "----------------------------"
./core/bin/apprun migrate status
echo ""

echo "Step 2: Reset database (drop all tables)"
echo "----------------------------------------"
./core/bin/apprun migrate reset
echo ""

echo "Step 3: Run sync (should apply pending migrations)"
echo "---------------------------------------------------"
./core/bin/apprun migrate sync
echo ""

echo "Step 4: Verify final status"
echo "---------------------------"
./core/bin/apprun migrate status
echo ""

echo "✅ Test completed successfully!"
echo ""
echo "Expected behavior:"
echo "  - Step 2: Database reset clears all tables"
echo "  - Step 3: Sync detects pending migrations and applies them"
echo "  - Step 4: Status shows 'No pending migrations'"
