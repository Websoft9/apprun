#!/bin/bash
# Auto-sync Story Index from sprint-status.yaml
# Generates story-index.md using the Python script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
STORY_INDEX="$PROJECT_ROOT/docs/sprint-artifacts/story-index.md"
SPRINT_STATUS="$PROJECT_ROOT/docs/sprint-artifacts/sprint-status.yaml"

# Check if sprint-status.yaml exists
if [ ! -f "$SPRINT_STATUS" ]; then
    echo "❌ sprint-status.yaml not found: $SPRINT_STATUS"
    exit 1
fi

echo "🔄 Syncing Story Index from sprint-status.yaml..."

# Generate story-index.md using Python script
if ! python3 "$SCRIPT_DIR/generate-story-index.py" --format markdown --output "$STORY_INDEX"; then
    echo "❌ Failed to generate story index"
    exit 1
fi

echo "✅ Story index synced: $STORY_INDEX"

# Show summary
if [ -f "$STORY_INDEX" ]; then
    total_stories=$(grep -c "^| [0-9]" "$STORY_INDEX" || echo 0)
    echo "📊 Total stories indexed: $total_stories"
    
    # Show status breakdown
    echo ""
    echo "📈 Status Breakdown:"
    grep "^- ✅" "$STORY_INDEX" || true
    grep "^- 🔄" "$STORY_INDEX" || true
    grep "^- 🟢" "$STORY_INDEX" || true
    grep "^- ✏️" "$STORY_INDEX" || true
    grep "^- 📋" "$STORY_INDEX" || true
fi

echo ""
echo "✨ Story index generation complete!"
